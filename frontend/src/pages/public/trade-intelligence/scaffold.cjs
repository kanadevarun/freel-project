const fs = require('fs');
const path = require('path');

const basePath = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/trade-intelligence';

const structure = {
  'guides': [
    { dir: 'incoterms', name: 'Incoterms' },
    { dir: 'air-freight', name: 'AirFreight' },
    { dir: 'sea-freight', name: 'SeaFreight' },
    { dir: 'customs-clearance', name: 'CustomsClearance' },
    { dir: 'documentation-guide', name: 'DocumentationGuide' },
    { dir: 'import-export-basics', name: 'ImportExportBasics' }
  ],
  'calculators': [
    { dir: 'cbm-calculator', name: 'CbmCalculator' },
    { dir: 'volumetric-weight', name: 'VolumetricWeight' },
    { dir: 'duty-calculator', name: 'DutyCalculator' },
    { dir: 'transit-time', name: 'TransitTime' },
    { dir: 'freight-cost', name: 'FreightCost' },
    { dir: 'container-load', name: 'ContainerLoad' }
  ],
  'references': [
    { dir: 'container-sizes', name: 'ContainerSizes' },
    { dir: 'port-directory', name: 'PortDirectory' },
    { dir: 'airport-directory', name: 'AirportDirectory' },
    { dir: 'hsn-codes', name: 'HsnCodes' },
    { dir: 'dangerous-goods', name: 'DangerousGoods' },
    { dir: 'country-trade-profiles', name: 'CountryTradeProfiles' }
  ],
  'insights': [
    { dir: 'logistics-trends', name: 'LogisticsTrends' },
    { dir: 'market-updates', name: 'MarketUpdates' },
    { dir: 'trade-news', name: 'TradeNews' },
    { dir: 'reports', name: 'Reports' },
    { dir: 'benchmarks', name: 'Benchmarks' },
    { dir: 'case-studies', name: 'CaseStudies' }
  ]
};

const sharedComponents = [
  'KnowledgeHero',
  'KnowledgeBreadcrumb',
  'KnowledgeSidebar',
  'KnowledgeCTA'
];

Object.keys(structure).forEach(section => {
  structure[section].forEach(page => {
    const dir = path.join(basePath, section, page.dir);
    fs.mkdirSync(dir, { recursive: true });
    
    // Page.jsx
    fs.writeFileSync(path.join(dir, 'Page.jsx'), 
`import './Page.css';

export default function ${page.name}Page() {
  return (
    <div className="ti-placeholder-page">
      <h1>${page.name} Guide Coming Soon</h1>
    </div>
  );
}
`);

    // Page.css
    fs.writeFileSync(path.join(dir, 'Page.css'), 
`.ti-placeholder-page {
  padding: 100px 24px;
  text-align: center;
}
`);
  });
});

// Shared
const sharedDir = path.join(basePath, 'shared');
fs.mkdirSync(sharedDir, { recursive: true });
sharedComponents.forEach(comp => {
  fs.writeFileSync(path.join(sharedDir, `${comp}.jsx`),
`import './${comp}.css';

export default function ${comp}() {
  return (
    <div className="ti-shared-comp">
      {/* ${comp} Content */}
    </div>
  );
}
`);
  fs.writeFileSync(path.join(sharedDir, `${comp}.css`), `/* ${comp} CSS */\n`);
});

// Rename existing ComingSoon.jsx and CSS
const oldCsJsx = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/TradeIntelligence/ComingSoon.jsx';
const oldCsCss = '/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/TradeIntelligence/ComingSoon.css';
const newCsDir = path.join(basePath, 'coming-soon');
fs.mkdirSync(newCsDir, { recursive: true });

if (fs.existsSync(oldCsJsx)) {
  const jsxContent = fs.readFileSync(oldCsJsx, 'utf-8');
  fs.writeFileSync(path.join(newCsDir, 'ComingSoonPage.jsx'), jsxContent.replace(/ComingSoon\.css/g, 'ComingSoonPage.css'));
  fs.unlinkSync(oldCsJsx);
}
if (fs.existsSync(oldCsCss)) {
  fs.copyFileSync(oldCsCss, path.join(newCsDir, 'ComingSoonPage.css'));
  fs.unlinkSync(oldCsCss);
}

// Ensure old TradeIntelligence folder is cleaned up if empty
try {
  fs.rmdirSync('/Users/varun.kanade/go/src/freel/freel-project/frontend/src/pages/public/TradeIntelligence');
} catch(e) {}

console.log("Scaffolding complete.");
