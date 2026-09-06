import { useState, useEffect, useRef } from 'react';
import PropTypes from 'prop-types';
import './LocationPickerMap.css';

/** Popular logistics ports / cities for quick selection */
const POPULAR_LOCATIONS = [
  'Mumbai, India',
  'Hamburg, Germany',
  'New York, USA',
  'Shanghai, China',
  'Dubai, UAE',
  'Singapore'
];

export default function LocationPickerMap({ value, onChange, placeholder = 'Search city, port, society, or address...', disabled = false, showMap = false }) {
  const [query, setQuery] = useState(value || '');
  const [selectedPlace, setSelectedPlace] = useState(value || '');
  const [suggestions, setSuggestions] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [isFocused, setIsFocused] = useState(false);

  const containerRef = useRef(null);
  const debounceTimerRef = useRef(null);

  useEffect(() => {
    setQuery(value || '');
    setSelectedPlace(value || '');
  }, [value]);

  // Handle clicking outside to close suggestions dropdown
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setShowDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Fetch live autocomplete suggestions as user types
  const fetchSuggestions = (inputText) => {
    if (!inputText || inputText.trim().length < 1) {
      setSuggestions([]);
      setShowDropdown(false);
      return;
    }

    setIsSearching(true);

    // 1. Try Google Maps AutocompleteService if window.google is loaded
    if (window.google && window.google.maps && window.google.maps.places) {
      try {
        const service = new window.google.maps.places.AutocompleteService();
        service.getPlacePredictions({ input: inputText }, (predictions, status) => {
          if (status === window.google.maps.places.PlacesServiceStatus.OK && predictions && predictions.length > 0) {
            const googleResults = predictions.map(p => p.description);
            setSuggestions(googleResults);
            setShowDropdown(true);
            setIsSearching(false);
            return;
          }
          // If AutocompleteService has no predictions, try Geocoding
          fetchMultiEngineFallback(inputText);
        });
        return;
      } catch (e) {
        console.warn('Google AutocompleteService warning, using fallback:', e);
      }
    }

    // 2. High-performance Multi-Engine Fallback (Photon + Nominatim) for societies, addresses & cities worldwide
    fetchMultiEngineFallback(inputText);
  };

  const fetchMultiEngineFallback = async (inputText) => {
    const searchTerms = inputText.trim();
    const encoded = encodeURIComponent(searchTerms);

    try {
      // Execute parallel queries against Photon (specializes in societies/buildings/streets) and Nominatim
      const [photonRes, nominatimRes] = await Promise.allSettled([
        fetch(`https://photon.komoot.io/api/?q=${encoded}&limit=6`),
        fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encoded}&limit=6`)
      ]);

      const resultsList = [];

      // Process Photon results
      if (photonRes.status === 'fulfilled' && photonRes.value.ok) {
        const photonData = await photonRes.value.json();
        if (photonData && photonData.features) {
          photonData.features.forEach(f => {
            const props = f.properties;
            const nameParts = [
              props.name,
              props.street || props.district || props.suburb,
              props.city || props.county || props.state,
              props.country
            ].filter(Boolean);
            if (nameParts.length > 0) {
              const fullStr = [...new Set(nameParts)].join(', ');
              resultsList.push(fullStr);
            }
          });
        }
      }

      // Process Nominatim results
      if (nominatimRes.status === 'fulfilled' && nominatimRes.value.ok) {
        const nominatimData = await nominatimRes.value.json();
        if (Array.isArray(nominatimData)) {
          nominatimData.forEach(item => {
            const display = item.display_name;
            const parts = display.split(', ');
            if (parts.length > 4) {
              resultsList.push(`${parts[0]}, ${parts[1]}, ${parts[parts.length - 2]}, ${parts[parts.length - 1]}`);
            } else {
              resultsList.push(display);
            }
          });
        }
      }

      // Deduplicate results
      const uniqueResults = [...new Set(resultsList)];

      if (uniqueResults.length > 0) {
        setSuggestions(uniqueResults);
        setShowDropdown(true);
      } else {
        // If exact society not found in remote db, offer typed value as direct suggestion option
        setSuggestions([searchTerms]);
        setShowDropdown(true);
      }
    } catch (err) {
      console.warn('Location suggestion query error:', err);
      setSuggestions([searchTerms]);
      setShowDropdown(true);
    } finally {
      setIsSearching(false);
    }
  };

  const handleInputChange = (e) => {
    const val = e.target.value;
    setQuery(val);
    onChange(val);

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    debounceTimerRef.current = setTimeout(() => {
      fetchSuggestions(val);
    }, 200); // Fast 200ms debounce
  };

  const handleSelectSuggestion = (placeStr) => {
    setQuery(placeStr);
    setSelectedPlace(placeStr);
    onChange(placeStr);
    setShowDropdown(false);
    setSuggestions([]);
  };

  const handleSelectQuickLocation = (loc) => {
    setQuery(loc);
    setSelectedPlace(loc);
    onChange(loc);
    setShowDropdown(false);
  };

  const handleClear = () => {
    setQuery('');
    setSelectedPlace('');
    setSuggestions([]);
    setShowDropdown(false);
    onChange('');
  };

  const encodeMapQuery = encodeURIComponent(selectedPlace || query || 'Mumbai, India');

  return (
    <div className="location-picker-wrapper" ref={containerRef}>
      {/* Input container with live autocomplete dropdown */}
      <div className="location-input-relative">
        <div className={`location-input-container ${isFocused ? 'focused' : ''}`}>
          <span className="location-search-icon">📍</span>
          <input
            type="text"
            className="location-search-input"
            value={query}
            onChange={handleInputChange}
            onFocus={() => {
              setIsFocused(true);
              if (suggestions.length > 0) setShowDropdown(true);
            }}
            onBlur={() => setIsFocused(false)}
            placeholder={placeholder}
            disabled={disabled}
            autoComplete="off"
          />
          {isSearching && <span className="location-spinner" />}
          {query && !isSearching && (
            <button type="button" className="location-clear-btn" onClick={handleClear} disabled={disabled} title="Clear location">
              ✕
            </button>
          )}
        </div>

        {/* Dynamic Autocomplete Suggestions Dropdown List */}
        {showDropdown && suggestions.length > 0 && (
          <ul className="location-suggestions-dropdown">
            {suggestions.map((item, idx) => (
              <li
                key={idx}
                className="suggestion-item"
                onMouseDown={(e) => {
                  e.preventDefault(); // Prevent blur before select
                  handleSelectSuggestion(item);
                }}
              >
                <span className="suggestion-icon">📍</span>
                <span className="suggestion-text">{item}</span>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Quick location chips */}
      <div className="location-quick-chips">
        <span className="quick-chip-label">Suggestions:</span>
        {POPULAR_LOCATIONS.map((loc) => (
          <button
            key={loc}
            type="button"
            className={`location-chip ${query === loc ? 'active' : ''}`}
            onClick={() => handleSelectQuickLocation(loc)}
            disabled={disabled}
          >
            {loc}
          </button>
        ))}
      </div>

      {/* Interactive Google Map preview widget (only if showMap is explicitly set to true) */}
      {showMap && (
        <div className="location-map-preview-card">
          <div className="map-card-header">
            <div className="map-card-title">
              <span className="map-title-icon">🗺️</span>
              <span>Geographic Map View</span>
            </div>
            {query && (
              <span className="map-location-tag" title={query}>
                <span className="live-dot" /> {query}
              </span>
            )}
          </div>
          <div className="map-iframe-container">
            <iframe
              title="Google Maps Location Preview"
              width="100%"
              height="180"
              style={{ border: 0, borderRadius: '8px' }}
              loading="lazy"
              allowFullScreen
              src={`https://maps.google.com/maps?q=${encodeMapQuery}&t=&z=14&ie=UTF8&iwloc=&output=embed`}
            />
          </div>
        </div>
      )}
    </div>
  );
}

LocationPickerMap.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  placeholder: PropTypes.string,
  disabled: PropTypes.bool,
  showMap: PropTypes.bool
};
