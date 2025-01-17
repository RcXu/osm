package ads

import (
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set"
	xds_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/openservicemesh/osm/pkg/envoy"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/metricsstore"
)

// getTypeResource invokes the XDS handler (LDS, CDS etc.) to respond to the XDS request containing the requests' type and associated resources
func (s *Server) getTypeResources(proxy *envoy.Proxy, request *xds_discovery.DiscoveryRequest) ([]types.Resource, error) {
	// Tracks the success of this TypeURI response operation; accounts also for receipt on envoy server side
	startedAt := time.Now()
	typeURI := envoy.TypeURI(request.TypeUrl)
	log.Trace().Str("proxy", proxy.String()).Msgf("Getting resources for type %s", typeURI.Short())

	handler, ok := s.xdsHandlers[typeURI]
	if !ok {
		return nil, errUnknownTypeURL
	}

	if s.catalog.GetMeshConfig().Spec.Observability.EnableDebugServer {
		s.trackXDSLog(proxy.GetName(), typeURI)
	}

	// Invoke XDS handler
	resources, err := handler(s.catalog, proxy, request, s.certManager, s.proxyRegistry)
	if err != nil {
		xdsPathTimeTrack(startedAt, typeURI, proxy, false)
		return nil, errCreatingResponse
	}

	xdsPathTimeTrack(startedAt, typeURI, proxy, true)
	return resources, nil
}

// sendResponse takes a set of TypeURIs which will be called to generate the xDS resources
// for, and will have them sent to the proxy server.
// If no DiscoveryRequest is passed, an empty one for the TypeURI is created
// TODO(draychev): Convert to variadic function: https://github.com/openservicemesh/osm/issues/3127
func (s *Server) sendResponse(proxy *envoy.Proxy, server *xds_discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer, request *xds_discovery.DiscoveryRequest, typeURIsToSend ...envoy.TypeURI) error {
	thereWereErrors := false

	// A nil request indicates a change on mesh configuration, OSM will trigger an update
	// for all proxy config (we generate a response with no direct request from envoy)
	osmDrivenUpdate := request == nil
	cacheResourceMap := map[string][]types.Resource{}

	// Order is important: CDS, EDS, LDS, RDS
	// See: https://github.com/envoyproxy/go-control-plane/issues/59
	for _, typeURI := range typeURIsToSend {
		// Handle request when is not provided, and the SDS case
		var finalReq *xds_discovery.DiscoveryRequest
		if osmDrivenUpdate {
			finalReq = &xds_discovery.DiscoveryRequest{TypeUrl: typeURI.String()}

			// Fill the request resources with subscribed resources from the proxy thus far.
			// Verticals should be held accountable for generating the requested resources.
			// Any additional resources generated by the verticals that have not been requested/subscribed to,
			// will be silently ignored by envoy.
			// For CDS and LDS, this is always an empty slice (wildcard)
			// For other verticals (RDS, EDS, SDS), this is the list of subscribed resources that we have last received
			// from the TypeURL at hand.
			finalReq.ResourceNames = getResourceSliceFromMapset(proxy.GetSubscribedResources(typeURI))
		} else {
			finalReq = request
		}

		// Generate the resources for this request
		resources, err := s.getTypeResources(proxy, finalReq)
		if err != nil {
			log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrGeneratingReqResource)).Str("proxy", proxy.String()).
				Msgf("Error generating response for typeURI: %s", typeURI.Short())
			thereWereErrors = true
			continue
		}

		if s.cacheEnabled {
			// Keep a reference to later set the full snapshot in the cache
			cacheResourceMap[typeURI.String()] = resources
		} else {
			// If cache disabled, craft and send a reply to the proxy on the stream
			if err := s.SendDiscoveryResponse(proxy, finalReq, server, resources); err != nil {
				log.Error().Err(err).Str("proxy", proxy.String()).Msgf("Error sending DiscoveryResponse for typeUrl: %s", typeURI.Short())
				thereWereErrors = true
			}
		}
	}

	if s.cacheEnabled {
		// Store the aggregated resources as a full snapshot
		if err := s.RecordFullSnapshot(proxy, cacheResourceMap); err != nil {
			log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrRecordingSnapshot)).Str("proxy", proxy.String()).
				Msgf("Error recording snapshot for proxy: %v", err)
			thereWereErrors = true
		}
	}

	isFullUpdate := len(typeURIsToSend) == len(envoy.XDSResponseOrder)
	if isFullUpdate {
		success := !thereWereErrors
		xdsPathTimeTrack(time.Now(), envoy.TypeADS, proxy, success)
	}

	return nil
}

// SendDiscoveryResponse creates a new response for <proxy> given <resourcesToSend> and <request.TypeURI> and sends it
func (s *Server) SendDiscoveryResponse(proxy *envoy.Proxy, request *xds_discovery.DiscoveryRequest, server *xds_discovery.AggregatedDiscoveryService_StreamAggregatedResourcesServer, resourcesToSend []types.Resource) error {
	// request.Node is only available on the first Discovery Request; will be nil on the following
	typeURI := envoy.TypeURI(request.TypeUrl)

	response := &xds_discovery.DiscoveryResponse{
		TypeUrl:     request.TypeUrl,
		VersionInfo: strconv.FormatUint(proxy.IncrementLastSentVersion(typeURI), 10),
		Nonce:       proxy.SetNewNonce(typeURI),
	}

	resourcesSent := mapset.NewSet()
	subscribedResources := proxy.GetSubscribedResources(typeURI)
	for _, res := range resourcesToSend {
		proto, err := anypb.New(res.(proto.Message))
		if err != nil {
			log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrMarshallingXDSResource)).
				Msgf("Error marshalling resource %s for proxy %s", typeURI, proxy.GetName())
			continue
		}
		// Add resource to response
		response.Resources = append(response.Resources, proto)

		// Only track as resources sent if they are subscribed resources.
		// By doing so, we are making sure a legitimate request down the line is not treated as an ACK just because
		// a vertical had potentially sent more resources when they had not been requested yet by the proxy.
		// For wildcard TypeURI resources, subscribed resources will purposefully remain empty at all times.
		if !envoy.IsWildcardTypeURI(typeURI) {
			currentResponseResourceName := cache.GetResourceName(res)
			if subscribedResources.Contains(currentResponseResourceName) {
				resourcesSent.Add(currentResponseResourceName)
			} else {
				log.Debug().Msgf("Proxy %s TypeURI %s - sending unsubscribed/unrequested resource %s",
					proxy.String(), typeURI.Short(), currentResponseResourceName)
			}
		}
	}

	// NOTE: Never log entire 'response' - will contain secrets!
	log.Trace().Msgf("Constructed %s response: VersionInfo=%s", response.TypeUrl, response.VersionInfo)

	// Validate the generated resources given the request
	validateRequestResponse(proxy, request, resourcesToSend)

	// Send the response
	if err := (*server).Send(response); err != nil {
		metricsstore.DefaultMetricsStore.ProxyResponseSendErrorCount.WithLabelValues(proxy.UUID.String(), proxy.Identity.String(), string(typeURI)).Inc()
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrSendingDiscoveryResponse)).
			Str("proxy", proxy.String()).Msgf("Error sending response for typeURI %s to proxy", typeURI.Short())
		return err
	}

	// Sending discovery response succeeded, record last resources sent
	proxy.SetLastResourcesSent(typeURI, resourcesSent)
	metricsstore.DefaultMetricsStore.ProxyResponseSendSuccessCount.WithLabelValues(proxy.UUID.String(), proxy.Identity.String(), string(typeURI)).Inc()

	return nil
}
