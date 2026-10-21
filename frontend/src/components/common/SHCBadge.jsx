import React from 'react';
import PropTypes from 'prop-types';
import {
  Flame,
  Thermometer,
  Shield,
  Anchor,
  AlertTriangle,
  Snowflake,
  Package
} from 'lucide-react';

/**
 * Standard International Special Handling Codes (SHC) used in Freight Forwarding
 * Compliant with IATA / FIATA / IMO Maritime & Multimodal Freight Standards.
 */
export const SHC_DEFINITIONS = {
  DGR: {
    code: 'DGR',
    shortLabel: 'DGR',
    title: 'Dangerous Goods',
    fullLabel: 'DGR · Dangerous Goods (HAZ)',
    description: 'Hazardous materials requiring IMO / IATA DG declaration, placarding, and stowage isolation.',
    Icon: Flame,
    bg: '#FEF2F2',
    border: '#FECACA',
    text: '#991B1B',
    dot: '#EF4444',
  },
  COL: {
    code: 'COL',
    shortLabel: 'COL',
    title: 'Temperature-Controlled',
    fullLabel: 'COL · Cold Chain / Reefer',
    description: 'Requires refrigerated container (reefer) with active continuous temperature monitoring.',
    Icon: Thermometer,
    bg: '#F0F9FF',
    border: '#BAE6FD',
    text: '#0369A1',
    dot: '#0EA5E9',
  },
  VAL: {
    code: 'VAL',
    shortLabel: 'VAL',
    title: 'Valuable Cargo',
    fullLabel: 'VAL · High Value Secured',
    description: 'High commercial value cargo requiring bonded vault custody, sealed GPS seal, and armed escort.',
    Icon: Shield,
    bg: '#FEFCE8',
    border: '#FEF08A',
    text: '#854D0E',
    dot: '#EAB308',
  },
  HEA: {
    code: 'HEA',
    shortLabel: 'HEA',
    title: 'Heavy / Out-of-Gauge',
    fullLabel: 'HEA · Heavy / OOG Cargo',
    description: 'Over-dimensional or heavy machinery requiring flat-rack, crane lift, or specialized lashings.',
    Icon: Anchor,
    bg: '#FAF5FF',
    border: '#E9D5FF',
    text: '#6B21A8',
    dot: '#A855F7',
  },
  FRG: {
    code: 'FRG',
    shortLabel: 'FRG',
    title: 'Fragile Cargo',
    fullLabel: 'FRG · Fragile / Delicate',
    description: 'High shock-sensitivity cargo requiring specialized top-load restriction and cushioned packaging.',
    Icon: AlertTriangle,
    bg: '#FFF7ED',
    border: '#FED7AA',
    text: '#9A3412',
    dot: '#F97316',
  },
  ICE: {
    code: 'ICE',
    shortLabel: 'ICE',
    title: 'Dry Ice',
    fullLabel: 'ICE · Dry Ice Refrigerant',
    description: 'Sublimating dry ice cooling requiring ventilation and non-pressurized cargo hold ventilation.',
    Icon: Snowflake,
    bg: '#ECFEFF',
    border: '#A5F3FC',
    text: '#155E75',
    dot: '#06B6D4',
  },
  GEN: {
    code: 'GEN',
    shortLabel: 'GEN',
    title: 'General Cargo',
    fullLabel: 'GEN · General Cargo',
    description: 'Standard ambient commercial cargo requiring standard dry van equipment without special permits.',
    Icon: Package,
    bg: '#F8FAFC',
    border: '#E2E8F0',
    text: '#475569',
    dot: '#94A3B8',
  },
};

/**
 * Intelligent detector to resolve SHC code from either:
 * 1. Explicit code ('DGR', 'COL', etc.)
 * 2. Object properties (shc, special_handling_code, is_dg, is_reefer, etc.)
 * 3. Text description regex search
 */
export function detectSHCCode(input) {
  if (!input) return 'GEN';

  if (typeof input === 'string') {
    const raw = input.toUpperCase().trim();
    if (SHC_DEFINITIONS[raw]) return raw;

    // Direct bracket tag: e.g. [SHC: DGR] or SHC: DGR
    const bracketMatch = raw.match(/SHC[:\s-]+([A-Z]{3})/);
    if (bracketMatch && SHC_DEFINITIONS[bracketMatch[1]]) {
      return bracketMatch[1];
    }

    if (/\b(DANGEROUS|HAZMAT|DG|FLAMMABLE|CHEMICAL|TOXIC|CORROSIVE|LITHIUM|EXPLOSIVE)\b/i.test(input)) {
      return 'DGR';
    }
    if (/\b(REEFER|COLD CHAIN|TEMP-CONTROLLED|TEMPERATURE CONTROLLED|PERISHABLE|CHILLED|FROZEN|PHARMA)\b/i.test(input)) {
      return 'COL';
    }
    if (/\b(VALUABLE|HIGH VALUE|BULLION|SECUR|LUXURY|JEWELRY)\b/i.test(input)) {
      return 'VAL';
    }
    if (/\b(HEAVY|OOG|OUT OF GAUGE|OVERSIZED|MACHINERY|CNC|ROBOTICS|MILLING|INDUSTRIAL EQUIPMENT|TRANSFORMER|GENERATOR)\b/i.test(input)) {
      return 'HEA';
    }
    if (/\b(FRAGILE|DELICATE|GLASSWARE|PRECISION INSTRUMENT|OPTICAL)\b/i.test(input)) {
      return 'FRG';
    }
    if (/\b(DRY ICE|CARBON DIOXIDE SOLID)\b/i.test(input)) {
      return 'ICE';
    }

    return 'GEN';
  }

  if (typeof input === 'object') {
    // Check explicit property
    const direct = input.special_handling_code || input.shc || input.special_handling;
    if (direct) {
      const resolved = detectSHCCode(direct);
      if (resolved !== 'GEN') return resolved;
    }

    // Check boolean flags
    if (input.is_dangerous_goods || input.is_dg || input.isDG) return 'DGR';
    if (input.is_temperature_controlled || input.is_reefer || input.isReefer || input.is_perishable) return 'COL';
    if (input.is_valuable) return 'VAL';
    if (input.is_heavy || input.is_oog || input.isOOG) return 'HEA';
    if (input.is_fragile) return 'FRG';

    // Check description or notes
    const textBlob = `${input.description || ''} ${input.cargo_description || ''} ${input.notes || ''} ${input.title || ''}`;
    if (textBlob.trim()) {
      return detectSHCCode(textBlob);
    }
  }

  return 'GEN';
}

/**
 * Visual Special Handling Code Badge component
 */
export default function SHCBadge({
  code,
  item,
  size = 'sm',
  variant = 'pill',
  showIcon = true,
  showCodeOnly = false,
  className = '',
  style = {}
}) {
  const resolvedCode = code ? detectSHCCode(code) : (item ? detectSHCCode(item) : 'GEN');
  const def = SHC_DEFINITIONS[resolvedCode] || SHC_DEFINITIONS.GEN;
  const { Icon } = def;

  const sizeStyles = {
    xs: {
      padding: '1px 6px',
      fontSize: '10px',
      gap: '3px',
      iconSize: 10,
      dotSize: 5,
    },
    sm: {
      padding: '2px 8px',
      fontSize: '11px',
      gap: '4px',
      iconSize: 12,
      dotSize: 6,
    },
    md: {
      padding: '4px 10px',
      fontSize: '12px',
      gap: '5.5px',
      iconSize: 14,
      dotSize: 7,
    },
  }[size] || sizeStyles.sm;

  const isCompact = variant === 'compact';

  return (
    <span
      className={`shc-badge shc-${resolvedCode.toLowerCase()} ${className}`}
      title={`${def.fullLabel}: ${def.description}`}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        fontWeight: 700,
        letterSpacing: '0.2px',
        lineHeight: 1.2,
        borderRadius: isCompact ? '4px' : '9999px',
        backgroundColor: def.bg,
        border: `1px solid ${def.border}`,
        color: def.text,
        padding: sizeStyles.padding,
        fontSize: sizeStyles.fontSize,
        gap: sizeStyles.gap,
        whiteSpace: 'nowrap',
        userSelect: 'none',
        verticalAlign: 'middle',
        boxShadow: '0 1px 2px rgba(0,0,0,0.03)',
        ...style,
      }}
    >
      {showIcon && <Icon size={sizeStyles.iconSize} style={{ flexShrink: 0, opacity: 0.9 }} />}
      <span>
        {showCodeOnly ? def.code : `SHC: ${def.code}`}
      </span>
      {!showCodeOnly && !isCompact && (
        <span style={{ fontWeight: 500, opacity: 0.85, marginLeft: '2px' }}>
          • {def.title}
        </span>
      )}
    </span>
  );
}

SHCBadge.propTypes = {
  code: PropTypes.string,
  item: PropTypes.any,
  size: PropTypes.oneOf(['xs', 'sm', 'md']),
  variant: PropTypes.oneOf(['pill', 'tag', 'subtle', 'compact']),
  showIcon: PropTypes.bool,
  showCodeOnly: PropTypes.bool,
  className: PropTypes.string,
  style: PropTypes.object,
};
