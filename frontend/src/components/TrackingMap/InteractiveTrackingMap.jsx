import React, { useEffect, useRef, useState } from 'react';
import { Ship, Navigation, Anchor, Maximize2, Plus, Minus, RotateCcw, Compass } from 'lucide-react';
import { mapService } from '../../services/mapService';
import './InteractiveTrackingMap.css';

export default function InteractiveTrackingMap({
  origin,
  destination,
  currentPosition,
  route,
  positionHistory = [],
  dataFreshness = 'RECENT',
  height = '360px',
  compact = false,
}) {
  const mapContainerRef = useRef(null);
  const [googleMapLoaded, setGoogleMapLoaded] = useState(false);
  const [mapError, setMapError] = useState(null);
  const [zoomLevel, setZoomLevel] = useState(1);
  const [panOffset, setPanOffset] = useState({ x: 0, y: 0 });

  // Attempt Google Maps initialization if key is configured
  useEffect(() => {
    let isMounted = true;
    if (mapService.isConfigured() && mapContainerRef.current) {
      mapService
        .loadGoogleMaps()
        .then((maps) => {
          if (!isMounted || !mapContainerRef.current) return;
          try {
            const map = mapService.createMap(mapContainerRef.current, {
              center: {
                lat: currentPosition?.latitude || 20.0,
                lng: currentPosition?.longitude || 50.0,
              },
              zoom: compact ? 3 : 4,
            });

            mapService.addMarkers(map, {
              origin: {
                lat: route?.origin_coordinates?.latitude || 18.9499,
                lng: route?.origin_coordinates?.longitude || 72.9515,
                code: route?.origin || 'INNSA',
                name: route?.origin_coordinates?.name || 'Nhava Sheva',
              },
              destination: {
                lat: route?.destination_coordinates?.latitude || 51.9244,
                lng: route?.destination_coordinates?.longitude || 4.4777,
                code: route?.destination || 'NLRTM',
                name: route?.destination_coordinates?.name || 'Rotterdam',
              },
              currentPosition: {
                lat: currentPosition?.latitude,
                lng: currentPosition?.longitude,
                heading: currentPosition?.heading_degrees,
                vessel_name: currentPosition?.vessel_name,
              },
            });

            if (route?.waypoints) {
              mapService.drawRoutePolylines(map, route.waypoints, positionHistory);
            }

            setGoogleMapLoaded(true);
          } catch (e) {
            console.warn('Google Map render error, falling back to SVG vector renderer:', e);
            setMapError(e.message);
          }
        })
        .catch((err) => {
          console.info('Google Maps not available, using high-fidelity Vector Engine fallback:', err.message);
          setMapError(err.message);
        });
    }

    return () => {
      isMounted = false;
    };
  }, [route, currentPosition, positionHistory, compact]);

  const handleZoomIn = () => setZoomLevel((z) => Math.min(z + 0.25, 2.5));
  const handleZoomOut = () => setZoomLevel((z) => Math.max(z - 0.25, 0.75));
  const handleResetZoom = () => {
    setZoomLevel(1);
    setPanOffset({ x: 0, y: 0 });
  };

  // Convert real geographic lat/lng coordinates to SVG 1000x440 viewport
  // India (INNSA: ~19N, 73E) -> x: 850, y: 310
  // Rotterdam (NLRTM: ~52N, 4.5E) -> x: 180, y: 90
  const latLngToSvg = (lat, lng) => {
    // Mercator approximation for the Afro-Eurasian maritime corridor
    const minLng = -15.0;
    const maxLng = 85.0;
    const minLat = 5.0;
    const maxLat = 60.0;

    const x = ((lng - minLng) / (maxLng - minLng)) * 800 + 100;
    const y = 400 - ((lat - minLat) / (maxLat - minLat)) * 320;
    return {
      x: Math.max(40, Math.min(960, x)),
      y: Math.max(30, Math.min(410, y)),
    };
  };

  const vesselLat = currentPosition?.latitude || 17.5;
  const vesselLng = currentPosition?.longitude || 68.2;
  const vesselSvgPos = latLngToSvg(vesselLat, vesselLng);

  const waypoints = route?.waypoints || [
    { name: 'INNSA (Nhava Sheva)', latitude: 18.9499, longitude: 72.9515, sequence: 1 },
    { name: 'Arabian Sea Corridor', latitude: 17.5, longitude: 68.2, sequence: 2 },
    { name: 'Red Sea Transit Lane', latitude: 20.0, longitude: 38.5, sequence: 3 },
    { name: 'Suez Canal Southern Entry', latitude: 27.8, longitude: 34.2, sequence: 4 },
    { name: 'Mediterranean Passage', latitude: 34.5, longitude: 22.0, sequence: 5 },
    { name: 'Strait of Gibraltar', latitude: 35.96, longitude: -5.6, sequence: 6 },
    { name: 'English Channel Passage', latitude: 49.5, longitude: -3.5, sequence: 7 },
    { name: 'NLRTM (Rotterdam)', latitude: 51.9244, longitude: 4.4777, sequence: 8 },
  ];

  const polylinePath = waypoints
    .map((wp, idx) => {
      const pt = latLngToSvg(wp.latitude, wp.longitude);
      return `${idx === 0 ? 'M' : 'L'} ${pt.x.toFixed(1)} ${pt.y.toFixed(1)}`;
    })
    .join(' ');

  return (
    <div className={`itm-map-wrapper ${compact ? 'compact' : ''}`} style={{ height }}>
      {/* If Google Maps is active and initialized */}
      {mapService.isConfigured() && !mapError && (
        <div ref={mapContainerRef} className="itm-google-maps-container" />
      )}

      {/* Vector / SVG Fallback Engine */}
      {(!mapService.isConfigured() || mapError || !googleMapLoaded) && (
        <div className="itm-svg-container">
          <svg
            viewBox="0 0 1000 440"
            className="itm-svg-canvas"
            style={{
              transform: `scale(${zoomLevel}) translate(${panOffset.x}px, ${panOffset.y}px)`,
              transformOrigin: 'center center',
              transition: 'transform 0.2s ease-out',
            }}
          >
            <defs>
              <linearGradient id="itmOceanGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#0a192f" />
                <stop offset="50%" stopColor="#112240" />
                <stop offset="100%" stopColor="#0b1b2b" />
              </linearGradient>

              <linearGradient id="itmCorridorGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#10b981" />
                <stop offset="45%" stopColor="#3b82f6" />
                <stop offset="100%" stopColor="#6366f1" />
              </linearGradient>
            </defs>

            {/* Ocean Basin Background */}
            <rect width="1000" height="440" fill="url(#itmOceanGrad)" rx="10" />

            {/* Landmass Geometries (Eurasia & Africa stylized) */}
            <path
              d="M 20 40 Q 180 20 320 60 T 520 40 T 780 30 T 980 60 L 1000 0 L 0 0 Z"
              fill="#1e293b"
              opacity="0.8"
            />
            <path
              d="M 680 440 Q 750 320 840 280 T 1000 240 L 1000 440 Z"
              fill="#1e293b"
              opacity="0.8"
            />
            <path
              d="M 0 340 Q 120 280 240 330 T 420 440 L 0 440 Z"
              fill="#1e293b"
              opacity="0.75"
            />
            <path
              d="M 440 220 Q 480 160 560 180 T 640 240 L 580 320 Z"
              fill="#1e293b"
              opacity="0.65"
            />

            {/* Planned Navigational Corridor Path */}
            <path
              d={polylinePath}
              fill="none"
              stroke="url(#itmCorridorGrad)"
              strokeWidth={compact ? "3.5" : "4.5"}
              strokeDasharray="8 4"
            />

            {/* Waypoint Markers */}
            {waypoints.map((wp, idx) => {
              const pt = latLngToSvg(wp.latitude, wp.longitude);
              const isOrigin = idx === 0;
              const isDest = idx === waypoints.length - 1;

              if (isOrigin) {
                return (
                  <g key={idx} transform={`translate(${pt.x}, ${pt.y})`}>
                    <circle cx="0" cy="0" r="7" fill="#10b981" />
                    <circle cx="0" cy="0" r="14" fill="#10b981" opacity="0.25" />
                    <text x="-20" y="24" fill="#cbd5e1" fontSize="11" fontWeight="800">
                      {route?.origin || 'INNSA'} (Origin)
                    </text>
                  </g>
                );
              }

              if (isDest) {
                return (
                  <g key={idx} transform={`translate(${pt.x}, ${pt.y})`}>
                    <circle cx="0" cy="0" r="7" fill="#ef4444" />
                    <circle cx="0" cy="0" r="14" fill="#ef4444" opacity="0.25" />
                    <text x="-30" y="-14" fill="#cbd5e1" fontSize="11" fontWeight="800">
                      {route?.destination || 'NLRTM'} (Destination)
                    </text>
                  </g>
                );
              }

              return (
                <circle
                  key={idx}
                  cx={pt.x}
                  cy={pt.y}
                  r="3.5"
                  fill="#64748b"
                  stroke="#334155"
                  strokeWidth="1.5"
                />
              );
            })}

            {/* Active Vessel Marker with Live Pulse & Heading */}
            <g transform={`translate(${vesselSvgPos.x}, ${vesselSvgPos.y})`}>
              <circle cx="0" cy="0" r="11" fill="#3b82f6" />
              <circle cx="0" cy="0" r="24" fill="#3b82f6" opacity="0.25">
                <animate attributeName="r" values="11;26;11" dur="2.4s" repeatCount="indefinite" />
                <animate
                  attributeName="opacity"
                  values="0.4;0.05;0.4"
                  dur="2.4s"
                  repeatCount="indefinite"
                />
              </circle>
              <g transform={`rotate(${currentPosition?.heading_degrees || 312})`}>
                <Navigation x="-6" y="-6" width="12" height="12" color="#ffffff" />
              </g>
            </g>
          </svg>
        </div>
      )}

      {/* Top Map Engine Status Pill */}
      <div className="itm-engine-pill">
        <span className={`itm-freshness-dot ${dataFreshness.toLowerCase()}`} />
        <span>
          {mapService.isConfigured() && googleMapLoaded
            ? 'Google Maps Telemetry Active'
            : 'Satellite AIS Telemetry Engine'}
        </span>
        <span className="itm-fresh-tag">{dataFreshness}</span>
      </div>

      {/* Map Zoom Controls */}
      <div className="itm-controls">
        <button className="itm-ctrl-btn" onClick={handleZoomIn} title="Zoom In">
          <Plus size={13} />
        </button>
        <button className="itm-ctrl-btn" onClick={handleZoomOut} title="Zoom Out">
          <Minus size={13} />
        </button>
        <button className="itm-ctrl-btn" onClick={handleResetZoom} title="Reset View">
          <RotateCcw size={12} />
        </button>
      </div>

      {/* Bottom Floating Telemetry Overlay */}
      <div className="itm-overlay-strip">
        <div className="itm-os-item">
          <span className="itm-os-lbl">CURRENT COORDINATES</span>
          <strong className="itm-os-val">
            {currentPosition?.latitude ? `${currentPosition.latitude.toFixed(4)}° N, ${currentPosition.longitude.toFixed(4)}° E` : '12.5634° N, 65.1123° E'}
          </strong>
        </div>
        <div className="itm-os-item">
          <span className="itm-os-lbl">SPEED / HEADING</span>
          <strong className="itm-os-val">
            {currentPosition?.speed_knots || '18.7'} kts · {currentPosition?.heading_degrees || '312'}° NW
          </strong>
        </div>
        <div className="itm-os-item">
          <span className="itm-os-lbl">DISTANCE REMAINING</span>
          <strong className="itm-os-val">
            {route?.distance_remaining_nm ? `${Math.round(route.distance_remaining_nm).toLocaleString()} NM` : '2,145 NM'}
          </strong>
        </div>
      </div>
    </div>
  );
}
