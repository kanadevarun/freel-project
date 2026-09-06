/**
 * MapService - Abstraction layer for Google Maps JavaScript API integration
 * Supports dynamic script injection, marker rendering, route polyline visualization,
 * and robust graceful fallback when no API key is present or when offline.
 */

let googleMapsPromise = null;

export const mapService = {
  /**
   * Check if Google Maps API key is configured in the environment.
   */
  isConfigured: () => {
    const key = import.meta.env.VITE_GOOGLE_MAPS_API_KEY;
    return Boolean(key && key.trim() !== '' && key !== 'YOUR_GOOGLE_MAPS_API_KEY');
  },

  /**
   * Dynamically loads the Google Maps JavaScript API script once.
   * Returns a promise that resolves when Google Maps is ready.
   */
  loadGoogleMaps: () => {
    if (typeof window === 'undefined') return Promise.reject(new Error('Window not available'));

    if (window.google && window.google.maps) {
      return Promise.resolve(window.google.maps);
    }

    if (googleMapsPromise) {
      return googleMapsPromise;
    }

    const apiKey = import.meta.env.VITE_GOOGLE_MAPS_API_KEY;
    if (!apiKey || apiKey.trim() === '' || apiKey === 'YOUR_GOOGLE_MAPS_API_KEY') {
      return Promise.reject(new Error('Google Maps API key is not configured'));
    }

    googleMapsPromise = new Promise((resolve, reject) => {
      const scriptId = 'google-maps-js-sdk';
      if (document.getElementById(scriptId)) {
        // Script already added, wait for window.google.maps
        const checkInterval = setInterval(() => {
          if (window.google && window.google.maps) {
            clearInterval(checkInterval);
            resolve(window.google.maps);
          }
        }, 100);
        return;
      }

      const script = document.createElement('script');
      script.id = scriptId;
      script.type = 'text/javascript';
      script.src = `https://maps.googleapis.com/maps/api/js?key=${encodeURIComponent(apiKey)}&libraries=geometry`;
      script.async = true;
      script.defer = true;

      script.onload = () => {
        if (window.google && window.google.maps) {
          resolve(window.google.maps);
        } else {
          reject(new Error('Google Maps SDK loaded but window.google.maps is undefined'));
        }
      };

      script.onerror = (err) => {
        googleMapsPromise = null;
        reject(new Error(`Failed to load Google Maps script: ${err?.message || 'Network error'}`));
      };

      document.head.appendChild(script);
    });

    return googleMapsPromise;
  },

  /**
   * Initializes a Google Map instance on a given DOM container.
   */
  createMap: (container, options = {}) => {
    if (!window.google || !window.google.maps) {
      throw new Error('Google Maps is not loaded');
    }

    const defaultOptions = {
      center: { lat: 25.0, lng: 45.0 },
      zoom: 3,
      mapTypeId: 'roadmap',
      disableDefaultUI: false,
      zoomControl: true,
      mapTypeControl: false,
      streetViewControl: false,
      fullscreenControl: true,
      styles: [
        { elementType: 'geometry', stylers: [{ color: '#1d2c4d' }] },
        { elementType: 'labels.text.fill', stylers: [{ color: '#8ec3b9' }] },
        { elementType: 'labels.text.stroke', stylers: [{ color: '#1a3646' }] },
        { featureType: 'water', elementType: 'geometry', stylers: [{ color: '#0e1626' }] },
        { featureType: 'water', elementType: 'labels.text.fill', stylers: [{ color: '#4e6d8d' }] },
      ],
      ...options,
    };

    return new window.google.maps.Map(container, defaultOptions);
  },

  /**
   * Adds custom styled markers for Origin, Destination, and Vessel.
   */
  addMarkers: (map, { origin, destination, currentPosition }) => {
    if (!window.google || !window.google.maps || !map) return [];
    const markers = [];

    // Origin Marker (Green)
    if (origin && origin.lat && origin.lng) {
      const originMarker = new window.google.maps.Marker({
        position: { lat: Number(origin.lat), lng: Number(origin.lng) },
        map,
        title: `Origin: ${origin.name || origin.code}`,
        icon: {
          path: window.google.maps.SymbolPath.CIRCLE,
          scale: 8,
          fillColor: '#10b981',
          fillOpacity: 1,
          strokeWeight: 2,
          strokeColor: '#ffffff',
        },
      });
      markers.push(originMarker);
    }

    // Destination Marker (Red)
    if (destination && destination.lat && destination.lng) {
      const destMarker = new window.google.maps.Marker({
        position: { lat: Number(destination.lat), lng: Number(destination.lng) },
        map,
        title: `Destination: ${destination.name || destination.code}`,
        icon: {
          path: window.google.maps.SymbolPath.CIRCLE,
          scale: 8,
          fillColor: '#ef4444',
          fillOpacity: 1,
          strokeWeight: 2,
          strokeColor: '#ffffff',
        },
      });
      markers.push(destMarker);
    }

    // Current Vessel Position (Blue with Heading Arrow)
    if (currentPosition && currentPosition.lat && currentPosition.lng) {
      const vesselMarker = new window.google.maps.Marker({
        position: { lat: Number(currentPosition.lat), lng: Number(currentPosition.lng) },
        map,
        title: `Current Vessel: ${currentPosition.vessel_name || 'Vessel in Transit'}`,
        icon: {
          path: window.google.maps.SymbolPath.FORWARD_CLOSED_ARROW,
          scale: 6,
          rotation: currentPosition.heading || 0,
          fillColor: '#3b82f6',
          fillOpacity: 1,
          strokeWeight: 2,
          strokeColor: '#ffffff',
        },
      });
      markers.push(vesselMarker);
    }

    return markers;
  },

  /**
   * Draws planned corridor and historical tracked route polylines.
   */
  drawRoutePolylines: (map, waypoints = [], historyPositions = []) => {
    if (!window.google || !window.google.maps || !map) return null;

    // 1. Planned Corridor Polyline (Dashed Blue)
    const plannedCoords = waypoints.map((w) => ({
      lat: Number(w.latitude || w.lat),
      lng: Number(w.longitude || w.lng),
    }));

    const plannedPolyline = new window.google.maps.Polyline({
      path: plannedCoords,
      geodesic: true,
      strokeColor: '#3b82f6',
      strokeOpacity: 0.8,
      strokeWeight: 3,
      map,
    });

    // 2. Historical Tracked Polyline (Solid Green)
    let historyPolyline = null;
    if (historyPositions.length > 1) {
      const historyCoords = historyPositions.map((h) => ({
        lat: Number(h.latitude || h.lat),
        lng: Number(h.longitude || h.lng),
      }));

      historyPolyline = new window.google.maps.Polyline({
        path: historyCoords,
        geodesic: true,
        strokeColor: '#10b981',
        strokeOpacity: 1.0,
        strokeWeight: 4,
        map,
      });
    }

    return { plannedPolyline, historyPolyline };
  },
};
