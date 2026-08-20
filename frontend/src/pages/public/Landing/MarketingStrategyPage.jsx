import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import './MarketingStrategy.css';

/* ════════════════════════════════════════════════════════════════
   LOGISTICSHQ — 30-Day Marketing Strategy Website
   ════════════════════════════════════════════════════════════════ */

/* ── DATA ── */
const PHASES = [
  { id: 1, days: 'Days 1–7', name: 'THE PROBLEM', goal: 'Make freight forwarders recognize themselves in the content.', tone: 'Empathetic, observational. No solutions. No product.', color: 'var(--phase1)', cls: 'p1' },
  { id: 2, days: 'Days 8–14', name: 'THE CHAOS', goal: 'Show the hidden complexity of a single RFQ workflow.', tone: 'Detailed, honest, slightly dry humour.', color: 'var(--phase2)', cls: 'p2' },
  { id: 3, days: 'Days 15–21', name: 'WHAT IF?', goal: 'Introduce possibility without hard-selling the product.', tone: 'Hopeful, speculative, question-driven.', color: 'var(--phase3)', cls: 'p3' },
  { id: 4, days: 'Days 22–26', name: 'THE REVEAL', goal: 'Clearly introduce LogisticsHQ features with honesty.', tone: 'Confident but grounded. Show real functionality.', color: 'var(--phase4)', cls: 'p4' },
  { id: 5, days: 'Days 27–30', name: 'SEE IT / TRY IT', goal: 'Convert awareness into demo bookings & early access.', tone: 'Direct. Urgency through scarcity.', color: 'var(--phase5)', cls: 'p5' },
];

const CALENDAR_DAYS = [
  { day: 1, phase: 1, platform: ['ig', 'li'], format: 'Reel + Text Post', title: '"POV: Customer just asked for a freight quote."', hook: 'A customer sends you this email. Watch what happens next.', content: 'Screen-recording style Reel showing the email → RFQ creation → missing weight → follow-up. LinkedIn: "The email that starts a shipment is almost never complete."', cta: 'Follow for more / Comment your experience', visual: 'Screen recording. Real email client. Dark UI. No stock photos.', lhq: false, stage: 'Awareness' },
  { day: 2, phase: 1, platform: ['ig'], format: 'Carousel (8 slides)', title: '"Why creating an RFQ from an email takes longer than it should"', hook: 'One customer email. Here\'s what your sales team actually does with it.', content: '8 slides: Sales reads email → opens RFQ form → copies details → realises weight missing → emails customer → waits 18 hours → customer replies → "There\'s got to be a better way."', cta: 'Save this if your team does this every day', visual: 'Clean white slides, dark text, emoji icons per step. Mobile legible.', lhq: false, stage: 'Problem recognition' },
  { day: 3, phase: 1, platform: ['yt', 'ig'], format: 'YouTube (8 min) + Static', title: '"Why Freight Forwarders Still Create RFQs Manually in 2025"', hook: 'A customer sends an email asking for a quote. What happens next is more complicated than you\'d think.', content: 'Full anatomy of a rate request email, the manual extraction problem, missing info loops, internal delays, cost of the delay. Instagram: "Your customer is waiting. Your sales team is waiting on the customer\'s weight."', cta: 'Subscribe / Tag someone who\'s been in this loop', visual: 'Screen recordings. Annotated walkthroughs. Founder voiceover.', lhq: false, stage: 'Problem recognition' },
  { day: 4, phase: 1, platform: ['li', 'ig'], format: 'Long-form Post + Reel', title: '"We mapped the journey from email to quotation. Here\'s where the delays happen."', hook: 'We sat with freight forwarding teams and timed each step. Uncomfortable results.', content: 'Step-by-step breakdown with time estimates: avg 2–3 days before pricing even starts. Instagram Reel: rapid-fire text overlays of the full workflow.', cta: 'Does this match your experience? / Follow if this is your daily reality', visual: 'LinkedIn: data-led text post. Instagram: black bg + white text overlays, fast cuts.', lhq: false, stage: 'Problem recognition' },
  { day: 5, phase: 1, platform: ['ig'], format: 'Carousel (10 slides)', title: '"The internal WhatsApp of every freight forwarding company:"', hook: 'The freight forwarding WhatsApp group everyone has and nobody talks about:', content: '10 slides styled exactly as WhatsApp messages. Fictional but hyper-realistic conversation between Sales, Pricing, MD, and the Customer following up.', cta: 'Save this. Send it to your WhatsApp group.', visual: 'Green WhatsApp bubbles, read receipts, timestamps. Instantly recognisable.', lhq: false, stage: 'Recognition' },
  { day: 5, phase: 1, platform: ['li'], format: 'Short observation', title: '"The average FF sales executive sends 3–5 internal messages before a customer gets a quote."', hook: 'The average freight forwarding sales executive sends 3–5 internal messages before a customer gets a quotation back.', content: 'Two paragraphs on the observation and its implication. Sales to Pricing. Pricing to Sales. Sales to Customer. Customer to Sales. Sales to Pricing again.', cta: 'How many steps does it take in your company?', visual: 'Pure text. First-person, conversational.', lhq: false, stage: 'Problem recognition' },
  { day: 6, phase: 1, platform: ['yt', 'ig'], format: 'YT Short + Static', title: '"How many tabs does your pricing team have open right now?"', hook: 'I asked a freight forwarding pricing manager to show me their screen. They had 7 tabs open.', content: 'Screen recording: 7 tabs — Maersk spot portal, MSC portal, Excel, Hapag-Lloyd rate file, email inbox, WhatsApp Web, PDF contract. All open for one RFQ. Instagram: infographic of the 7 tabs.', cta: 'Subscribe. We\'re about to go deep on this. / Comment the number.', visual: 'Real browser tabs. Minimal commentary. Lets the visual speak.', lhq: false, stage: 'Problem recognition' },
  { day: 7, phase: 1, platform: ['li', 'ig'], format: 'Doc Post + Reel', title: '"One RFQ. 12 steps. Zero automation."', hook: 'This is the full internal workflow that happens every time a customer asks for a freight quote.', content: 'LinkedIn: 12-slide document post. One step per slide with time estimates and human cost. Final: "Total time: 2–4 days. Automated steps: 0." Instagram: Launch of Rate Desk Reality series.', cta: 'Which step takes longest? / Follow + Save. New episode every week.', visual: 'LinkedIn: dark minimal slides. Instagram: series title card, premium branding.', lhq: false, stage: 'Problem recognition' },
  { day: 8, phase: 2, platform: ['yt', 'li'], format: 'YouTube (10 min) + Text Post', title: '"Why Carrier Contract PDFs Are So Difficult to Work With"', hook: 'A freight forwarder receives a new carrier contract. It\'s 47 pages. Here\'s what they actually need to do with it.', content: 'Inside a carrier contract: note quotes, surcharge codes (BAF, PSS, OHC, DHC), multiple rate structures per lane, validity windows. LinkedIn: "Pricing isn\'t a rate lookup problem. It\'s an information integration problem."', cta: 'Subscribe / Comment: biggest pricing team pain?', visual: 'Real-style PDF walkthrough. Annotations. Callouts. Professional.', lhq: false, stage: 'Education' },
  { day: 9, phase: 2, platform: ['ig'], format: 'Carousel (8 slides)', title: '"What\'s actually inside a carrier contract PDF?"', hook: 'Your carrier just sent you a 47-page contract. Here\'s what\'s actually inside it.', content: 'Rate tables, note quotes, surcharge schedules, validity windows, port code variations, different formats per carrier. Final slide: "Your pricing team reads every one of these manually."', cta: 'Save this. Send to your pricing team.', visual: 'Recreated PDF-style slides. Authentic visual language.', lhq: false, stage: 'Education' },
  { day: 10, phase: 2, platform: ['li', 'ig'], format: 'Text Post + Reel', title: '"Contract rates and spot rates are not the same thing."', hook: 'Contract rates and spot rates are not the same thing. Most software treats them as if they are.', content: 'Full breakdown of contract vs spot rates, when each applies, surcharges, why same lane can have 3 different valid rates. Instagram: Rate Desk Reality Episode 1 — pricing manager\'s screen walkthrough.', cta: 'What\'s the bigger pain: contract or spot? / Rate Desk Reality — follow for weekly episodes.', visual: 'LinkedIn: expert voice. Instagram: screen recording, authentic rate search workflow.', lhq: false, stage: 'Education' },
  { day: 11, phase: 2, platform: ['ig', 'li'], format: 'Static Post + Short', title: '"Contract Rate vs Spot Rate: The decision for every RFQ."', hook: 'Contract rate vs Spot rate: The decision your pricing team makes for every single RFQ.', content: 'Two-column infographic: Contract rate (negotiated annually, carrier PDF, different format, may be expired) vs Spot rate (current market, carrier portal, changes daily, may miss surcharges).', cta: 'How does your team decide? / What information is your team spending time chasing?', visual: 'Clean infographic. Dark bg. Two columns. Highly shareable.', lhq: false, stage: 'Education' },
  { day: 12, phase: 2, platform: ['yt'], format: 'YouTube (8 min)', title: '"Contract Rates vs Spot Rates: What Your Pricing Team Actually Uses"', hook: 'When a customer asks for a freight quote, your pricing team has to answer a question harder than it looks: Which rate do I use?', content: 'Full breakdown: contract vs spot, when each applies, how surcharges change the picture, same lane with 3 different valid rates. First subtle hint: "we\'re building something to help."', cta: 'Next week: When AI can actually help. Subscribe.', visual: 'Structured explainer. Diagrams. Real rate examples (anonymized).', lhq: 'hint', stage: 'Education' },
  { day: 13, phase: 2, platform: ['ig', 'li'], format: 'Carousel + Short', title: '"The Sales → Pricing handoff. Here\'s every way it goes wrong."', hook: 'Sales creates the RFQ. Pricing prices it. Sounds simple. Here\'s what actually happens.', content: '10-slide carousel: RFQ arrives incomplete → Pricing asks for missing info → Sales goes to customer → 2-day delay → Pricing asks which contract → Sales doesn\'t know → everyone waits → "Everyone is working. Nobody has the same information."', cta: 'Save this. Which slide is your company stuck on? / What would one place for all information mean?', visual: 'Clean, numbered slides. Escalating frustration in visual language.', lhq: false, stage: 'Problem recognition' },
  { day: 14, phase: 2, platform: ['li', 'ig'], format: 'Doc Post + Poll', title: '"The true cost of fragmented rate data in freight forwarding."', hook: 'Rate data is not the problem. Rate data fragmentation is.', content: 'LinkedIn: Document post — cost examples (time per RFQ, missed lanes, expiry errors, inconsistent surcharges). Final: "This is a solvable problem." Instagram: Stories poll on biggest daily pain points.', cta: 'Is your rate data management working? / Poll results shared next day.', visual: 'LinkedIn: authoritative data document. Instagram: polls, interactive.', lhq: false, stage: 'Education' },
  { day: 15, phase: 3, platform: ['li', 'ig'], format: 'Text Post + Reel', title: '"What if the customer\'s email could become the RFQ automatically?"', hook: 'What if the customer\'s email could become the RFQ automatically?', content: 'First "what if" post. Not "we built this" — just exploring the idea. Email contains 6 of 8 fields needed. What if AI read it? What could be detected as missing? Instagram: Side-by-side animation — email fields mapping to RFQ fields.', cta: 'Would you trust AI to do this? / Follow — we\'re building this.', visual: 'LinkedIn: speculative, intellectual. Instagram: clean animation, field-by-field reveal.', lhq: 'subtle', stage: 'Curiosity' },
  { day: 16, phase: 3, platform: ['yt'], format: 'YouTube (9 min)', title: '"How AI Could Help Freight Forwarding Pricing Teams (And Where It Shouldn\'t)"', hook: 'AI is being applied to logistics. But most implementations miss the most painful part: the human work before a shipment is booked.', content: 'What AI could help: RFQ creation, rate search, rule application. What AI should NOT do alone: approve quotes, pricing strategy, handle exceptions. The human-in-the-loop model. What a well-designed system looks like.', cta: 'Next week: We show you what we\'ve built. Subscribe.', visual: 'Thoughtful, structured. Two-column thinking — AI tasks vs human tasks.', lhq: 'yes', stage: 'Curiosity' },
  { day: 17, phase: 3, platform: ['ig', 'li'], format: 'Carousel + Text', title: '"8 tasks that could be automated — and 4 that absolutely shouldn\'t be."', hook: 'Not everything in freight forwarding should be automated. But some things definitely should.', content: 'Could automate: email→RFQ, missing info detection, rate search, surcharge calculation. Should NOT automate: pricing approval, contract negotiation, customer relationships. "The best systems know the difference."', cta: 'Save this. Tag your pricing manager. / Where would you draw the line?', visual: 'Split visual: green checkmarks vs red flags. Clear, scannable.', lhq: 'yes', stage: 'Curiosity' },
  { day: 18, phase: 3, platform: ['ig', 'li'], format: 'Reel + Text Post', title: '"What if contract and spot rates could be searched at the same time?"', hook: 'What if contract rates and spot rates could be searched simultaneously?', content: 'Animation: carrier portal + PDF "worlds" merge into one interface. Both rates side by side, surcharges normalised, margin calculated. LinkedIn: First honest product-building post — the contract normalization engineering challenge.', cta: 'This is what we\'re building. Follow @logisticshq. / Interested in seeing this? DM.', visual: 'Instagram: clean animation, two systems merging. LinkedIn: technical-lite, honest.', lhq: 'yes', stage: 'Curiosity' },
  { day: 19, phase: 3, platform: ['yt'], format: 'Founder Video (6 min)', title: '"Why We\'re Building LogisticsHQ"', hook: 'I\'ve spent 6 months sitting with freight forwarding teams, watching how they work. Here\'s what I saw.', content: 'Founder story. Personal, authentic. The observations. The conversations. The moment of realisation. Specific problems that led to building. NOT a product pitch — a story about why.', cta: 'Subscribe. Next week we show you what we\'ve built.', visual: 'Founder on camera. Natural light. Minimal set. Authentic over polished.', lhq: 'yes', stage: 'Trust' },
  { day: 20, phase: 3, platform: ['ig'], format: 'Carousel (7 slides)', title: '"What if pricing rules could be applied automatically?"', hook: 'Your pricing team has rules. You just haven\'t written them down in a system yet.', content: '"Minimum margin: 8%." "INNSA-DEHAM lane promo: add 12%." "If buy > $8,000 — flag for review." What if a system applied these automatically? And flagged exceptions for human review? "That\'s not removing the pricing manager. It\'s giving them their time back."', cta: 'This is what we\'re building at LogisticsHQ. Link in bio to join early access.', visual: 'Dark slides. Pricing rules shown as code/config blocks. Premium feel.', lhq: 'yes', stage: 'Product curiosity' },
  { day: 21, phase: 3, platform: ['li'], format: 'Text Post', title: '"The future of freight forwarding isn\'t removing people from the workflow."', hook: 'The future of freight forwarding isn\'t removing people from the workflow. It\'s removing repetitive work from the workflow.', content: 'Full philosophy statement. Repetitive work: email parsing, data extraction, rate lookup, margin calculation. Human work: judgment, relationships, exceptions, decisions. "Every piece of software we\'re building distinguishes between these two categories."', cta: 'If you\'re a freight forwarding professional interested in early access, comment \'interested\' or DM.', visual: 'Pure text. Founder voice. Definitive, memorable.', lhq: 'yes', stage: 'Trust' },
  { day: 22, phase: 4, platform: ['yt', 'ig'], format: 'Full Demo (10 min) + Reel', title: '"LogisticsHQ — From Customer Email to Quotation. Full Workflow."', hook: 'In the last 3 weeks, we\'ve shown you how broken the freight forwarding workflow is. Today we show you what we\'ve built to fix it.', content: 'Full end-to-end demo. Real RFQ: email parsing, missing field detection, pricing agent, contract+spot rates, pricing rules, margin validation, human review, quote approval. Instagram: Fast 60-sec product teaser — 6-8 second shots of each feature.', cta: 'Link in description to book demo or early access.', visual: 'YouTube: professional demo recording. Instagram: premium product shots, music + text overlays.', lhq: true, stage: 'Product discovery' },
  { day: 23, phase: 4, platform: ['li', 'ig'], format: 'Doc Post + Carousel', title: '"LogisticsHQ: What we\'ve built and why."', hook: 'For 3 weeks, we\'ve been talking about what\'s broken in freight forwarding. Today we show you what we\'ve built.', content: 'LinkedIn: Document post — problem (1 slide), then one feature per slide in business terms (not technical). Final: early access CTA. Instagram: Sales Workspace carousel — email → RFQ flow with product screenshots.', cta: 'DM for personalised demo. / Early access open. Link in bio.', visual: 'LinkedIn: clean dark doc slides. Instagram: product UI screenshots, premium styling.', lhq: true, stage: 'Product discovery' },
  { day: 24, phase: 4, platform: ['yt', 'li'], format: 'Feature Walkthrough (7 min) + Post', title: '"LogisticsHQ Pricing Agent — Contract Rates, Spot Rates, and Human Review"', hook: 'Here\'s exactly how LogisticsHQ searches contract and spot rates simultaneously — and what happens when the AI finds an anomaly.', content: 'Screen recording: pricing workspace, rate comparison view, AI activity log, anomaly flag, human review modal, approve/edit/reject. LinkedIn: "Our AI pricing agent prepares the quote but asks a human before sending. Here\'s why."', cta: 'Book a demo — link in description. / Would you trust AI to prepare quotes?', visual: 'Clean product screen recording. Annotated UI elements. Professional pacing.', lhq: true, stage: 'Product education' },
  { day: 25, phase: 4, platform: ['ig', 'li'], format: 'Before/After Reel + Post', title: '"Before LogisticsHQ. After LogisticsHQ. Same RFQ."', hook: 'Before LogisticsHQ. After LogisticsHQ. Same RFQ. Same team. Very different timeline.', content: 'Split screen: Before (12-step manual, 3 days), After (AI agent, 4 steps, 10 minutes, human approval). LinkedIn: "We just shipped Contract Intelligence. Here\'s what it actually does." Technical-lite explanation.', cta: 'Early access. Link in bio. / Want to see it with your contracts? DM.', visual: 'Split screen with time counters. Premium animation. LinkedIn: honest, engineering-adjacent voice.', lhq: true, stage: 'Product education' },
  { day: 26, phase: 4, platform: ['yt'], format: 'Product Walkthrough (8 min)', title: '"LogisticsHQ Contract Intelligence — How We Process Carrier PDFs"', hook: 'You\'ve just received a 47-page carrier contract. Here\'s what LogisticsHQ does with it.', content: 'Full demo: contract upload → AI processing → extracted rates → human validation → rates available in pricing workspace. Shows the exact workflow that removes the manual PDF search.', cta: 'Book a demo. Link in description.', visual: 'Full product screen recording. Real contract-style PDF. Professional voice.', lhq: true, stage: 'Product education' },
  { day: 27, phase: 5, platform: ['yt', 'li'], format: 'End-to-End Demo (12 min) + Campaign Close', title: '"Mumbai → Hamburg. Full Freight Forwarding Workflow in LogisticsHQ."', hook: 'Let\'s run a real RFQ. Customer requests a rate from Mumbai to Hamburg, 3×40HC, FOB. Watch what happens.', content: 'Real demo. Customer email → AI reads → creates RFQ → detects missing HS code → pricing agent runs → Maersk, MSC, CMA CGM rates → pricing rules → margin validated → human review → quote ready. LinkedIn: Campaign retrospective post with early access CTA.', cta: 'Book your demo — early access seats are limited. / Comment \'demo\' or DM.', visual: 'YouTube: professional demo. LinkedIn: warm, personal retrospective.', lhq: true, stage: 'Conversion' },
  { day: 28, phase: 5, platform: ['ig', 'li'], format: 'Founder Reel + Direct CTA Post', title: '"Why we spent 6 months talking to freight forwarders before writing a line of code."', hook: 'Why we spent 6 months talking to freight forwarders before writing a single line of code.', content: 'Founder video — personal, authentic. Conversations that shaped the product. The moment they decided to build. LinkedIn: Direct early access invitation. "We\'d like to show you a 30-minute personalised demo using your actual lanes and carriers."', cta: 'Early access open. Link in bio. / Book directly: [link]. Or DM.', visual: 'Instagram: founder on camera, natural. LinkedIn: warm, direct, personal.', lhq: true, stage: 'Trust + Conversion' },
  { day: 29, phase: 5, platform: ['yt', 'li'], format: 'Short + Doc Post', title: '"The one thing that surprised everyone we showed LogisticsHQ to."', hook: 'We showed this to 10 freight forwarding pricing managers. Every single one said the same thing.', content: 'Focus on AI activity timeline — the moment AI shows every step it took, then asks human to review a specific anomaly. "This is exactly what I\'ve been asking for." LinkedIn: Document post — what beta users said, real observations.', cta: 'See it yourself. Link in bio / description.', visual: 'Short: 60s product moment. LinkedIn: authentic beta feedback presentation.', lhq: true, stage: 'Trust' },
  { day: 30, phase: 5, platform: ['ig', 'li', 'yt'], format: 'Campaign Close — All Platforms', title: '"30 days of freight forwarding problems. One solution."', hook: '30 days. One story. Here\'s everything we covered — and what comes next.', content: 'Instagram: Story series recapping journey, poll "Are you ready to see it?", strong CTA. LinkedIn: Full retrospective post. YouTube: 5-min compilation video — clips from all 4 weeks building the narrative arc. Ends: "LogisticsHQ. Early access open."', cta: 'Book your demo. Early access: 20 freight forwarding companies in first cohort.', visual: 'All formats reprise the campaign visual language. Cohesive retrospective.', lhq: true, stage: 'Conversion' },
];

const SERIES = [
  { num: '01', title: 'Rate Desk Reality', formats: ['Instagram Reels', 'YouTube Shorts'], freq: 'Weekly (Fridays)', audience: 'Pricing managers, pricing executives', purpose: 'Show the real complexity of freight pricing without sensationalism', episodes: ['What happens when a carrier contract PDF arrives', 'Finding a note quote in a 47-page document', 'When spot and contract rates disagree', 'Calculating surcharges — BAF, PSS, OHC', 'The validity date problem', 'What AI-assisted pricing actually looks like'] },
  { num: '02', title: 'One RFQ, 10 Problems', formats: ['Instagram Carousels', 'LinkedIn Docs'], freq: 'Bi-weekly', audience: 'FF owners, Sales managers', purpose: 'Break down a realistic RFQ scenario and show every friction point', episodes: ['Mumbai → Hamburg. Electronics. 3×40HC.', 'Delhi → Singapore. Air. 800kg. One missing HS code.', 'JNPT → Rotterdam. DG Cargo. The compliance chain.'] },
  { num: '03', title: 'Freight Forwarding Explained', formats: ['YouTube Long-form'], freq: 'Weekly', audience: 'Anyone learning about freight forwarding', purpose: 'SEO authority, educational positioning, top-of-funnel', episodes: ['What is a Freight Forwarder? The Real Role', 'FCL vs LCL: How Cargo Gets Consolidated', 'Incoterms Explained — EXW to DDP', 'What is a Bill of Lading?', 'How Ocean Freight Rates Are Actually Set'] },
  { num: '04', title: 'Would You Automate This?', formats: ['Instagram Stories', 'LinkedIn Questions'], freq: 'Weekly (Mondays)', audience: 'All FF personas', purpose: 'Engagement, community, gauge comfort with automation', episodes: ['Reading the customer email and creating the RFQ?', 'Detecting missing shipment information?', 'Calculating the margin on a quote?', 'Approving the final quote?', 'Searching rates but not approving them?'] },
  { num: '05', title: "Inside a FF's Inbox", formats: ['Instagram Carousels'], freq: 'Weekly', audience: 'Sales executives, FF owners', purpose: 'Humorous recognition, deep empathy, high shareability', episodes: ['A normal Monday morning inbox', 'Every type of customer email your sales team receives', 'The 6 carrier emails you didn\'t ask for', 'The rate request that became a 3-day project'] },
  { num: '06', title: 'Before / After LogisticsHQ', formats: ['Instagram Reels', 'LinkedIn Posts'], freq: 'Weekly (Phases 4–5)', audience: 'All personas', purpose: 'Product demonstration through narrative comparison', episodes: ['Creating an RFQ: Before vs After', 'Searching for rates: Before vs After', 'Quote approval: Before vs After', 'Full workflow: Before vs After'] },
];

const FUNNEL_STEPS = [
  { stage: 'AWARENESS', title: 'Problem Storytelling Content', desc: 'Instagram Reels, LinkedIn observations, YouTube explainers showing the freight forwarding workflow pain without mentioning LogisticsHQ.', platforms: ['ig', 'li', 'yt'] },
  { stage: 'RECOGNITION', title: 'Website: Homepage Hero', desc: '"Freight forwarding wasn\'t supposed to be this complicated." Visitors from social land here and immediately recognise their problem.', platforms: ['web'] },
  { stage: 'CURIOSITY', title: '"What If?" Content + Product Teasers', desc: 'Phase 3 content introduces possibility. Links begin driving to product feature pages.', platforms: ['ig', 'li', 'yt'] },
  { stage: 'EDUCATION', title: 'Website: Product + Feature Pages', desc: 'Sales Workspace, Pricing Workspace, Contract Intelligence, AI Agent pages explain what LogisticsHQ does in the context of known problems.', platforms: ['web'] },
  { stage: 'DISCOVERY', title: 'Product Demos + Walkthroughs', desc: 'YouTube demos, Instagram Before/After, LinkedIn product posts drive directly to demo page.', platforms: ['ig', 'li', 'yt'] },
  { stage: 'CONVERSION', title: 'Demo Booking / Early Access', desc: '"Book a Demo" page with personalised offer. "Early access: 20 FF companies in first cohort."', platforms: ['web'] },
  { stage: 'TRUST', title: 'Sales Conversation → Onboarding', desc: 'Lead captured, qualified, and converted to early access customer through personal demo.', platforms: ['web'] },
];

const KPI_DATA = {
  instagram: [
    { metric: 'Reel views', vals: ['2,000+', '5,000+', '10,000+', '15,000+'], pcts: [14, 33, 67, 100] },
    { metric: 'Followers gained', vals: ['50+', '100+', '150+', '200+'], pcts: [25, 50, 75, 100] },
    { metric: 'Carousel saves', vals: ['50+', '100+', '150+', '200+'], pcts: [25, 50, 75, 100] },
    { metric: 'Link-in-bio clicks', vals: ['20+', '40+', '100+', '200+'], pcts: [10, 20, 50, 100] },
  ],
  linkedin: [
    { metric: 'Impressions', vals: ['2K+', '4K+', '8K+', '12K+'], pcts: [17, 33, 67, 100] },
    { metric: 'Engagement rate', vals: ['3%+', '4%+', '5%+', '5%+'], pcts: [60, 80, 100, 100] },
    { metric: 'Quality comments', vals: ['10+', '20+', '30+', '40+'], pcts: [25, 50, 75, 100] },
    { metric: 'Demo interest', vals: ['2+', '5+', '10+', '20+'], pcts: [10, 25, 50, 100] },
  ],
  youtube: [
    { metric: 'Click-through rate', vals: ['4%+', '5%+', '6%+', '7%+'], pcts: [57, 71, 86, 100] },
    { metric: 'Avg watch time', vals: ['50%+', '55%+', '55%+', '60%+'], pcts: [83, 92, 92, 100] },
    { metric: 'New subscribers', vals: ['20+', '40+', '60+', '100+'], pcts: [20, 40, 60, 100] },
    { metric: 'Website clicks', vals: ['10+', '20+', '40+', '80+'], pcts: [13, 25, 50, 100] },
  ],
  business: [
    { metric: 'Website visitors', vals: ['100+', '300+', '600+', '1,000+'], pcts: [10, 30, 60, 100] },
    { metric: 'Demo requests', vals: ['2+', '5+', '10+', '20+'], pcts: [10, 25, 50, 100] },
    { metric: 'Early access signups', vals: ['5+', '15+', '30+', '50+'], pcts: [10, 30, 60, 100] },
    { metric: 'Qualified FF leads', vals: ['1+', '3+', '7+', '15+'], pcts: [7, 20, 47, 100] },
  ],
};

const CHECKLIST_ITEMS = {
  prelaunch: [
    'Instagram account created, bio updated, link-in-bio set up',
    'LinkedIn company page + founder personal profile optimised',
    'YouTube channel created with branding, description, links',
    'Website /logisticshq page live and optimised',
    '/contact demo booking form ready and tested',
    'Early access waitlist form live (Typeform or native)',
    'Brand kit finalised: colours, fonts, logo, icon',
    '30-day content calendar shared with the full team',
    'First week of content shot/designed and scheduled',
    'Screen recording setup ready for product demos',
    'Founder prepared and comfortable on camera',
    'Scheduling tool set up (Buffer, Later, or Hootsuite)',
  ],
  week1: [
    'Reel 1: POV customer email (45-min shoot)',
    'Carousel 1: "One email, what happens next" (2h design)',
    'Static posts x2 (1h design each)',
    'LinkedIn posts x5 (30 min writing each)',
    'YouTube video 1: RFQ manual process (half day)',
    'Stories x7 (daily, 10 min each)',
    'Thumbnail designed for YouTube video 1',
    'Caption + hashtag research completed',
  ],
  week2: [
    'Rate Desk Reality Episode 1 (screen recording, 1h)',
    'Carousel: Inside a carrier contract PDF (2h design)',
    'YouTube video 2: Carrier contract PDFs (half day)',
    'YouTube video 3: Contract vs Spot rates (half day)',
    'LinkedIn document posts x1 (design + write, 2h)',
    '"Would You Automate This?" Stories x2 (15 min each)',
    'Stories poll results analysed and published',
  ],
  week3: [
    'Product screenshots captured — all major features',
    'Founder video recorded (Day 19) — 2h session',
    'Animation: Email → RFQ field mapping (1 day)',
    '"What if?" Reel series — 3 reels (half day each)',
    'LinkedIn long-form posts x3 (1h each)',
    'Early access page / waitlist form live',
    'Beta user feedback collected and documented',
  ],
  week4: [
    'Full product demo video recorded (Day 22) — 2h',
    'Pricing Agent walkthrough (Day 24) — 1h',
    'Contract Intelligence demo (Day 26) — 1h',
    'Mumbai→Hamburg full demo (Day 27) — 2h',
    'Before/After Reel (Day 25) — 2h shoot',
    'Campaign retrospective video (Day 30) — 2h',
    'Week 4 KPIs reviewed and documented',
    'Month 2 campaign planning session scheduled',
  ],
};

const WEEK_SCHEDULE = [
  { day: 'Monday', items: [['📝', 'LinkedIn long-form thought leadership post'], ['📊', 'Instagram "Would You Automate This?" Story poll'], ['🗓', 'Schedule week\'s content in tool'], ['💬', 'Respond to previous week\'s comments']] },
  { day: 'Tuesday', items: [['🎬', 'Instagram Reel published'], ['✏️', 'LinkedIn short observation post'], ['👂', 'Respond to Monday\'s comments and DMs']] },
  { day: 'Wednesday', items: [['📸', 'Instagram Carousel published'], ['▶️', 'YouTube video published (if ready)'], ['📄', 'LinkedIn workflow breakdown or product post'], ['📊', 'Mid-week Stories: quick insight or tip']] },
  { day: 'Thursday', items: [['🖼', 'Instagram Static post'], ['📑', 'LinkedIn Document post (bi-weekly)'], ['⚡', 'YouTube Short (repurposed Reel)'], ['📊', 'Review mid-week metrics']] },
  { day: 'Friday', items: [['🎬', 'Rate Desk Reality episode (Instagram)'], ['🔍', 'LinkedIn end-of-week observation'], ['📈', 'Weekly KPI review session'], ['📅', 'Next week content planning (30 min)']] },
];

/* ── Hooks ── */
function useReveal() {
  useEffect(() => {
    const timer = setTimeout(() => {
      const els = document.querySelectorAll('.ms-reveal');
      const obs = new IntersectionObserver(
        (entries) => entries.forEach(e => { if (e.isIntersecting) { e.target.classList.add('visible'); obs.unobserve(e.target); } }),
        { threshold: 0.08, rootMargin: '0px 0px -30px 0px' }
      );
      els.forEach(el => obs.observe(el));
      return () => obs.disconnect();
    }, 100);
    return () => clearTimeout(timer);
  }, []);
}

/* ── Navbar ── */
function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  useEffect(() => {
    const h = () => setScrolled(window.scrollY > 40);
    window.addEventListener('scroll', h, { passive: true });
    return () => window.removeEventListener('scroll', h);
  }, []);
  const go = (id) => document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });
  return (
    <nav className={`ms-navbar${scrolled ? ' scrolled' : ''}`} id="ms-navbar">
      <div className="ms-nav-logo">
        <div className="ms-nav-logo-icon">⚡</div>
        LogisticsHQ
      </div>
      <div className="ms-nav-links">
        <button className="ms-nav-btn" onClick={() => go('ms-phases')}>Phases</button>
        <button className="ms-nav-btn" onClick={() => go('ms-calendar')}>Calendar</button>
        <button className="ms-nav-btn" onClick={() => go('ms-platforms')}>Platforms</button>
        <button className="ms-nav-btn" onClick={() => go('ms-series')}>Series</button>
        <button className="ms-nav-btn" onClick={() => go('ms-kpis')}>KPIs</button>
        <Link to="/logisticshq" className="ms-nav-cta" id="ms-nav-view-website">View Website →</Link>
      </div>
    </nav>
  );
}

/* ── Hero ── */
function Hero() {
  const northStarSteps = [
    { text: '"Freight forwarding is chaotic."', hi: false },
    { text: '"Why is it still this manual?"', hi: false },
    { text: '"Someone should solve this."', hi: false },
    { text: '"Maybe AI can actually help here."', hi: false },
    { text: '"LogisticsHQ is solving this."', hi: true },
    { text: '"I want to see how it works."', hi: false },
    { text: '"I want early access."', hi: true },
  ];
  return (
    <section className="ms-hero" id="ms-hero">
      <div className="ms-hero-bg">
        <div className="ms-hero-glow-1" />
        <div className="ms-hero-glow-2" />
        <div className="ms-hero-grid" />
      </div>
      <div className="ms-hero-inner">
        <div className="ms-hero-badge">
          <div className="ms-hero-badge-dot" />
          30-Day Marketing & Content Launch Strategy
        </div>
        <h1 className="ms-hero-title">
          Build the story first.<br />
          <span className="ms-hero-accent">Sell LogisticsHQ second.</span>
        </h1>
        <p className="ms-hero-sub">
          A complete content campaign for Instagram, YouTube, and LinkedIn — designed to make freight forwarders recognise their problems before we reveal the solution.
        </p>

        <div className="ms-hero-north-star">
          {northStarSteps.map((s, i) => (
            <div key={i}>
              <div className="ms-ns-step">
                <span className={`ms-ns-text${s.hi ? ' highlight' : ''}`}>{s.text}</span>
              </div>
              {i < northStarSteps.length - 1 && <div className="ms-ns-arrow" />}
            </div>
          ))}
        </div>

        <div className="ms-hero-cta-row">
          <button className="ms-btn-primary" onClick={() => document.getElementById('ms-calendar')?.scrollIntoView({ behavior: 'smooth' })} id="ms-hero-view-calendar">
            View 30-Day Calendar
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
          </button>
          <button className="ms-btn-secondary" onClick={() => document.getElementById('ms-platforms')?.scrollIntoView({ behavior: 'smooth' })} id="ms-hero-view-platforms">
            Platform Strategy
          </button>
        </div>
      </div>
    </section>
  );
}

/* ── Phases ── */
function Phases() {
  return (
    <section className="ms-section ms-section-light" id="ms-phases">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Campaign Structure</div>
          <h2 className="ms-section-title">5 Narrative Phases. 30 Days.</h2>
          <p className="ms-section-sub">Each phase has a specific emotional goal. We earn the audience's trust before we ask for their attention.</p>
        </div>
        <div className="ms-phases-row">
          {PHASES.map((p, i) => (
            <div key={i} className={`ms-reveal ms-phase-card ${p.cls}`} style={{ transitionDelay: `${i * 0.1}s` }}>
              <div className="ms-phase-days">{p.days}</div>
              <div className="ms-phase-name">{p.name}</div>
              <div className="ms-phase-goal">{p.goal}</div>
              <div className="ms-phase-tone">{p.tone}</div>
              <div className="ms-phase-big-num">{p.id}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ── 30-Day Calendar ── */
function Calendar() {
  const [filter, setFilter] = useState('all');
  const [selected, setSelected] = useState(null);

  const filtered = filter === 'all'
    ? CALENDAR_DAYS
    : CALENDAR_DAYS.filter(d => d.platform.includes(filter));

  const allDayNums = [...new Set(CALENDAR_DAYS.map(d => d.day))];

  const phaseClass = (phase) => `phase-${phase}`;

  const platformLabel = (p) => {
    if (p === 'ig') return 'IG';
    if (p === 'li') return 'LI';
    if (p === 'yt') return 'YT';
    if (p === 'all') return 'ALL';
    return p;
  };

  const closeModal = useCallback((e) => {
    if (e.key === 'Escape') setSelected(null);
  }, []);
  useEffect(() => {
    window.addEventListener('keydown', closeModal);
    return () => window.removeEventListener('keydown', closeModal);
  }, [closeModal]);

  return (
    <section className="ms-calendar-section" id="ms-calendar">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Execution Plan</div>
          <h2 className="ms-section-title">30-Day Content Calendar</h2>
          <p className="ms-section-sub">Every day: platform, format, hook, content, CTA, and visual direction. Click any day for full details.</p>
        </div>

        <div className="ms-calendar-filter ms-reveal ms-d1">
          {[['all', 'All Days'], ['ig', '📸 Instagram'], ['li', '💼 LinkedIn'], ['yt', '▶️ YouTube']].map(([val, label]) => (
            <button key={val} className={`ms-filter-btn f-${val}${filter === val ? ' active' : ''}`} onClick={() => setFilter(val)}>{label}</button>
          ))}
          <div style={{ marginLeft: 'auto', display: 'flex', gap: '6px', alignItems: 'center', flexWrap: 'wrap' }}>
            {PHASES.map(p => (
              <span key={p.id} style={{ display: 'inline-flex', alignItems: 'center', gap: '5px', fontSize: '11px', color: 'var(--ms-text-3)' }}>
                <span style={{ width: '8px', height: '8px', borderRadius: '2px', background: p.color, display: 'inline-block' }} />
                {p.name}
              </span>
            ))}
          </div>
        </div>

        <div className="ms-reveal ms-d2 ms-calendar-grid">
          {allDayNums.map(dayNum => {
            const entries = CALENDAR_DAYS.filter(d => d.day === dayNum);
            const first = entries[0];
            const isDimmed = filter !== 'all' && !entries.some(e => e.platform.includes(filter));
            const allPlatforms = [...new Set(entries.flatMap(e => e.platform))];
            return (
              <div
                key={dayNum}
                className={`ms-cal-day ${phaseClass(first.phase)}${isDimmed ? ' dimmed' : ''}`}
                onClick={() => !isDimmed && setSelected(first)}
                role="button"
                tabIndex={0}
                aria-label={`Day ${dayNum}: ${first.title}`}
                onKeyDown={e => e.key === 'Enter' && !isDimmed && setSelected(first)}
              >
                <div className="ms-cal-day-num">DAY {dayNum}</div>
                <div className="ms-cal-platform-row">
                  {allPlatforms.map(p => <span key={p} className={`ms-cal-plat ${p}`}>{platformLabel(p)}</span>)}
                </div>
                <div className="ms-cal-title">{first.title}</div>
                <div className="ms-cal-format">{first.format}</div>
              </div>
            );
          })}
        </div>

        {selected && (
          <div className="ms-cal-modal-overlay" onClick={(e) => e.target === e.currentTarget && setSelected(null)}>
            <div className="ms-cal-modal">
              <button className="ms-modal-close" onClick={() => setSelected(null)} aria-label="Close">✕</button>
              <div className="ms-modal-day-num">DAY {selected.day} · PHASE {selected.phase} · {PHASES[selected.phase - 1].name}</div>
              <div className="ms-modal-title">{selected.title}</div>

              <div className="ms-modal-hook">"{selected.hook}"</div>

              <div className="ms-modal-grid">
                <div className="ms-modal-field">
                  <div className="ms-modal-field-label">Platform</div>
                  <div style={{ display: 'flex', gap: '6px' }}>
                    {selected.platform.map(p => <span key={p} className={`ms-cal-plat ${p}`}>{platformLabel(p)}</span>)}
                  </div>
                </div>
                <div className="ms-modal-field">
                  <div className="ms-modal-field-label">Format</div>
                  <div className="ms-modal-field-val">{selected.format}</div>
                </div>
                <div className="ms-modal-field">
                  <div className="ms-modal-field-label">Funnel Stage</div>
                  <div className="ms-modal-field-val">{selected.stage}</div>
                </div>
                <div className="ms-modal-field">
                  <div className="ms-modal-field-label">Mention LogisticsHQ?</div>
                  <div>
                    <span className={`ms-modal-mention ${selected.lhq === true ? 'yes' : 'no'}`}>
                      {selected.lhq === true ? '✓ Yes' : selected.lhq === 'yes' ? '✓ Yes' : selected.lhq === 'subtle' ? '◎ Subtle' : selected.lhq === 'hint' ? '◎ Hint only' : '✗ No'}
                    </span>
                  </div>
                </div>
                <div className="ms-modal-field full">
                  <div className="ms-modal-field-label">Content Concept</div>
                  <div className="ms-modal-field-val">{selected.content}</div>
                </div>
                <div className="ms-modal-field full">
                  <div className="ms-modal-field-label">Visual Direction</div>
                  <div className="ms-modal-field-val">{selected.visual}</div>
                </div>
                <div className="ms-modal-cta-box" style={{ gridColumn: '1/-1' }}>
                  <div className="ms-modal-field-label" style={{ marginBottom: '4px' }}>Call to Action</div>
                  <div className="ms-modal-field-val" style={{ color: 'var(--ms-amber)', fontSize: '13px', fontWeight: 600 }}>{selected.cta}</div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}

/* ── Brand Story ── */
function BrandStory() {
  const beliefs = [
    'Freight forwarding professionals are not inefficient. Their tools are.',
    "AI shouldn't replace the pricing manager. It should give them their time back.",
    "The problem isn't that people work slowly. Information lives in too many places.",
    "Most logistics software digitises the transaction. Very little addresses the work before it.",
    "The future of FF isn't removing people from the workflow. It's removing repetitive work.",
  ];
  const dont = ["AI-powered logistics platform", "Revolutionise your freight forwarding", "Transform your supply chain", "Unlock the power of AI", "Next-generation logistics"];
  const doSay = [
    "Your pricing team doesn't need another spreadsheet.",
    "Why are we still creating RFQs from emails manually in 2025?",
    "One customer email creates an entire workflow nobody asked for.",
    "Your sales team isn't slow. Your workflow is.",
    "Pricing isn't just finding the cheapest rate.",
    "Everyone has the information. Nobody has it in the same place.",
  ];
  return (
    <section className="ms-section ms-section-surface" id="ms-brand">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Brand Narrative</div>
          <h2 className="ms-section-title">What we say. What we never say.</h2>
          <p className="ms-section-sub">The campaign must feel like an industry movement — not a generic SaaS advertisement.</p>
        </div>

        <div className="ms-beliefs-grid ms-reveal ms-d1">
          {beliefs.map((b, i) => (
            <div key={i} className="ms-belief-card">
              <div className="ms-belief-num">0{i + 1}</div>
              <div>{b}</div>
            </div>
          ))}
        </div>

        <div className="ms-story-grid ms-reveal ms-d2" style={{ marginTop: '32px' }}>
          <div className="ms-story-col dont">
            <div className="ms-story-col-title">❌ Never Say</div>
            {dont.map((t, i) => <div key={i} className="ms-story-item"><span>✗</span><span>"{t}"</span></div>)}
          </div>
          <div className="ms-story-col do">
            <div className="ms-story-col-title">✅ Say Instead</div>
            {doSay.map((t, i) => <div key={i} className="ms-story-item"><span style={{ color: 'var(--ms-teal)' }}>✓</span><span>"{t}"</span></div>)}
          </div>
        </div>
      </div>
    </section>
  );
}

/* ── Platform Strategies ── */
function Platforms() {
  const plats = [
    {
      cls: 'ms-plat-ig', icon: '📸', name: 'Instagram',
      role: 'Emotional + visual storytelling. First point of awareness for pricing and sales professionals.',
      formats: [['Reels (30–60s)', '4–5 / week'], ['Carousels (6–10 slides)', '2–3 / week'], ['Static posts', '2–3 / week'], ['Stories', 'Daily']],
      note: 'Every Reel must hook in the first 1.5 seconds. Every Carousel Slide 1 is a provocation, not an explanation. Always end problem content without a solution — build tension.',
    },
    {
      cls: 'ms-plat-yt', icon: '▶️', name: 'YouTube',
      role: 'Educational authority. Long-form trust building for FF founders, owners, and operations managers.',
      formats: [['Long-form explainers', '1–2 / week'], ['Product walkthroughs', 'Phase 4–5 only'], ['Founder story videos', '1 per phase'], ['YouTube Shorts', '3–4 / week']],
      note: 'Channel positioning: "The channel that actually explains freight forwarding — and what\'s broken about it." SEO targets: freight forwarding software, RFQ workflow, carrier rate intelligence.',
    },
    {
      cls: 'ms-plat-li', icon: '💼', name: 'LinkedIn',
      role: 'Industry authority and B2B trust. Builds founder credibility and decision-maker reach.',
      formats: [['Long-form text posts', '2–3 / week'], ['Document posts', '1 / week'], ['Short observations', '2 / week'], ['Video posts (repurposed)', '1 / week']],
      note: 'LinkedIn is NOT Instagram repurposed. Categories: founder observations, workflow breakdowns, data/insight posts, product-building stories, contrarian takes on AI/logistics.',
    },
  ];
  return (
    <section className="ms-section ms-section-light" id="ms-platforms">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Platform Playbooks</div>
          <h2 className="ms-section-title">Three platforms. Three strategies.</h2>
          <p className="ms-section-sub">Each platform serves a different role in the funnel. They complement — never simply duplicate — each other.</p>
        </div>
        <div className="ms-platforms-grid">
          {plats.map((p, i) => (
            <div key={i} className={`ms-reveal ms-plat-card ${p.cls}`} style={{ transitionDelay: `${i * 0.12}s` }}>
              <div className="ms-plat-icon">{p.icon}</div>
              <div className="ms-plat-name">{p.name}</div>
              <div className="ms-plat-role">{p.role}</div>
              <div className="ms-plat-formats">
                {p.formats.map(([name, freq], j) => (
                  <div key={j} className="ms-plat-format-row">
                    <span className="ms-plat-format-name">{name}</span>
                    <span className="ms-plat-format-freq">{freq}</span>
                  </div>
                ))}
              </div>
              <div className="ms-plat-note">{p.note}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ── Content Series ── */
function ContentSeries() {
  return (
    <section className="ms-section ms-section-surface" id="ms-series">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Recurring Series</div>
          <h2 className="ms-section-title">6 branded content series.</h2>
          <p className="ms-section-sub">Recurring series build audience expectation and brand recall — people come back for the next episode.</p>
        </div>
        <div className="ms-series-grid">
          {SERIES.map((s, i) => (
            <div key={i} className={`ms-reveal ms-series-card`} style={{ transitionDelay: `${i * 0.08}s` }}>
              <div className="ms-series-num">Series {s.num}</div>
              <div className="ms-series-title">"{s.title}"</div>
              <div className="ms-series-meta">
                {s.formats.map((f, j) => <span key={j} className="ms-series-tag">{f}</span>)}
                <span className="ms-series-tag">📅 {s.freq}</span>
              </div>
              <div className="ms-series-purpose">{s.purpose}</div>
              <div className="ms-series-episodes">
                <div style={{ fontSize: '11px', fontWeight: 700, color: 'var(--ms-text-3)', letterSpacing: '1px', textTransform: 'uppercase', marginBottom: '4px' }}>Episodes</div>
                {s.episodes.map((ep, j) => (
                  <div key={j} className="ms-series-ep">
                    <span className="ms-series-ep-dot">›</span>
                    <span>{ep}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ── Content Funnel ── */
function ContentFunnel() {
  return (
    <section className="ms-section ms-section-light" id="ms-funnel">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Full Funnel</div>
          <h2 className="ms-section-title">From first scroll to first demo.</h2>
          <p className="ms-section-sub">Every piece of content maps to a specific stage. Every stage has a specific destination.</p>
        </div>
        <div className="ms-funnel-container ms-reveal ms-d1">
          {FUNNEL_STEPS.map((step, i) => (
            <div key={i} className="ms-funnel-step">
              <div className="ms-funnel-spine">
                <div className="ms-funnel-dot" />
                <div className={`ms-funnel-line${i === FUNNEL_STEPS.length - 1 ? ' last' : ''}`} />
              </div>
              <div className="ms-funnel-content">
                <div className="ms-funnel-stage">{step.stage}</div>
                <div className="ms-funnel-title">{step.title}</div>
                <div className="ms-funnel-desc">{step.desc}</div>
                <div className="ms-funnel-platforms">
                  {step.platforms.map(p => <span key={p} className={`ms-funnel-plat ${p}`}>{p === 'ig' ? '📸 Instagram' : p === 'li' ? '💼 LinkedIn' : p === 'yt' ? '▶️ YouTube' : '🌐 Website'}</span>)}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ── CTA Strategy ── */
function CTAStrategy() {
  const phases = [
    { days: '1–7', cta: '"Follow for more"', sec: '"Save this" / "Tag your team"', tone: 'Build audience, no ask yet' },
    { days: '8–14', cta: '"Subscribe" (YouTube)', sec: '"Comment your experience"', tone: 'Deepen relationship' },
    { days: '15–21', cta: '"Join our waitlist"', sec: '"Follow @logisticshq"', tone: 'Begin conversion journey' },
    { days: '22–26', cta: '"Book a demo"', sec: '"Early access — link in bio"', tone: 'Direct conversion push' },
    { days: '27–30', cta: '"Book your demo (20 spots)"', sec: '"DM me directly"', tone: 'Urgency + scarcity' },
  ];
  return (
    <section className="ms-section ms-section-surface" id="ms-cta-strategy">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">CTA Strategy</div>
          <h2 className="ms-section-title">We earn the ask. Then we make it.</h2>
          <p className="ms-section-sub">CTA intensity increases with audience trust. We never ask for a demo before we've earned the right to.</p>
        </div>
        <div className="ms-reveal ms-d1" style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {phases.map((p, i) => (
            <div key={i} style={{ display: 'grid', gridTemplateColumns: '80px 1fr 1fr 1fr', gap: '16px', alignItems: 'center', padding: '16px 20px', background: 'var(--ms-card)', border: '1px solid var(--ms-border)', borderRadius: '12px' }}>
              <div style={{ fontFamily: 'var(--ms-mono)', fontSize: '11px', color: 'var(--ms-text-3)' }}>Days {p.days}</div>
              <div style={{ fontSize: '13px', fontWeight: 700, color: 'var(--ms-teal)' }}>{p.cta}</div>
              <div style={{ fontSize: '13px', color: 'var(--ms-text-2)' }}>{p.sec}</div>
              <div style={{ fontSize: '12px', color: 'var(--ms-text-3)', fontStyle: 'italic' }}>{p.tone}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ── Posting Frequency ── */
function PostingFrequency() {
  const freqs = [
    {
      platform: 'Instagram', icon: '📸',
      items: [['Reels', '4–5 / week'], ['Carousels', '2–3 / week'], ['Static posts', '2–3 / week'], ['Stories', 'Daily']],
      reason: 'Reels are the discovery engine on Instagram. 4–5/week is aggressive but necessary for a launch phase. Daily Stories maintain a warm audience without heavy production.',
    },
    {
      platform: 'LinkedIn', icon: '💼',
      items: [['Long-form text posts', '2–3 / week'], ['Short observations', '2 / week'], ['Document posts', '1 / week'], ['Video posts', '1 / week']],
      reason: "LinkedIn's algorithm rewards consistency. 5 posts/week = one per weekday. This is a B2B platform — consistency over virality. Quality of engagement matters more than reach.",
    },
    {
      platform: 'YouTube', icon: '▶️',
      items: [['Long-form videos', '1–2 / week'], ['YouTube Shorts', '3–4 / week'], ['Founder videos', '1 per phase'], ['Product demos', 'Phase 4–5 only']],
      reason: 'Long-form YouTube takes 3–7 days to index and begin ranking. Start early, post consistently. 2 long-form/week is the maximum sustainable rate without quality drop. Shorts extend reach.',
    },
  ];
  return (
    <section className="ms-section ms-section-light" id="ms-frequency">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Posting Frequency</div>
          <h2 className="ms-section-title">Realistic. Deliberate. Consistent.</h2>
          <p className="ms-section-sub">These numbers aren't arbitrary — each one is justified by platform algorithm behaviour and team capacity.</p>
        </div>
        <div className="ms-freq-grid">
          {freqs.map((f, i) => (
            <div key={i} className={`ms-reveal ms-freq-card`} style={{ transitionDelay: `${i * 0.12}s` }}>
              <div className="ms-freq-platform">
                <span className="ms-freq-platform-icon">{f.icon}</span>
                {f.platform}
              </div>
              <div className="ms-freq-items">
                {f.items.map(([name, freq], j) => (
                  <div key={j} className="ms-freq-item">
                    <span className="ms-freq-format">{name}</span>
                    <span className="ms-freq-num">{freq}</span>
                  </div>
                ))}
              </div>
              <div className="ms-freq-reason">{f.reason}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ── KPI Dashboard ── */
function KPIDashboard() {
  const [tab, setTab] = useState('instagram');
  const tabs = [['instagram', '📸 Instagram'], ['linkedin', '💼 LinkedIn'], ['youtube', '▶️ YouTube'], ['business', '📊 Business']];
  const data = KPI_DATA[tab];
  return (
    <section className="ms-section ms-section-surface" id="ms-kpis">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Measurement Framework</div>
          <h2 className="ms-section-title">Weekly KPIs by platform.</h2>
          <p className="ms-section-sub">Measured every Friday. Adjustments made every Monday. The key early signal is saves and quality comments — not just reach.</p>
        </div>
        <div className="ms-reveal ms-d1">
          <div className="ms-kpi-tabs">
            {tabs.map(([val, label]) => (
              <button key={val} className={`ms-kpi-tab${tab === val ? ' active' : ''}`} onClick={() => setTab(val)}>{label}</button>
            ))}
          </div>
          <div className="ms-kpi-grid">
            {data.map((kpi, i) => (
              <div key={i} className="ms-kpi-card">
                <div className="ms-kpi-metric">{kpi.metric}</div>
                <div className="ms-kpi-weeks">
                  {[0, 1, 2, 3].map(w => (
                    <div key={w} className="ms-kpi-week-row">
                      <span className="ms-kpi-week-label">W{w + 1}</span>
                      <div className="ms-kpi-week-bar-wrap">
                        <div className="ms-kpi-week-bar" style={{ width: `${kpi.pcts[w]}%` }} />
                      </div>
                      <span className="ms-kpi-week-val">{kpi.vals[w]}</span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

/* ── Weekly Schedule ── */
function WeeklySchedule() {
  return (
    <section className="ms-section ms-section-light" id="ms-schedule">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Weekly Rhythm</div>
          <h2 className="ms-section-title">What the team does each day.</h2>
          <p className="ms-section-sub">A repeatable weekly cadence that the team can follow without a daily decision about what to post.</p>
        </div>
        <div className="ms-week-grid">
          {WEEK_SCHEDULE.map((day, i) => (
            <div key={i} className={`ms-reveal ms-week-day-card`} style={{ transitionDelay: `${i * 0.1}s` }}>
              <div className="ms-week-day-name">{day.day}</div>
              <div className="ms-week-items">
                {day.items.map(([icon, text], j) => (
                  <div key={j} className="ms-week-item">
                    <span className="ms-week-item-icon">{icon}</span>
                    <span>{text}</span>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ── Production Checklist ── */
function ProductionChecklist() {
  const [activeWeek, setActiveWeek] = useState('prelaunch');
  const [checked, setChecked] = useState({});
  const weeks = [['prelaunch', 'Pre-Launch'], ['week1', 'Week 1'], ['week2', 'Week 2'], ['week3', 'Week 3'], ['week4', 'Week 4']];

  const toggle = (key) => setChecked(prev => ({ ...prev, [key]: !prev[key] }));

  const items = CHECKLIST_ITEMS[activeWeek] || [];
  const doneCount = items.filter((_, i) => checked[`${activeWeek}-${i}`]).length;

  return (
    <section className="ms-section ms-section-surface" id="ms-checklist">
      <div className="ms-container">
        <div className="ms-section-header ms-reveal">
          <div className="ms-label">Production Checklist</div>
          <h2 className="ms-section-title">What needs to be produced.</h2>
          <p className="ms-section-sub">Click items to mark them complete. Track progress week by week.</p>
        </div>
        <div className="ms-reveal ms-d1">
          <div className="ms-checklist-tabs">
            {weeks.map(([val, label]) => (
              <button key={val} className={`ms-checklist-tab${activeWeek === val ? ' active' : ''}`} onClick={() => setActiveWeek(val)}>{label}</button>
            ))}
            <div style={{ marginLeft: 'auto', fontSize: '13px', color: 'var(--ms-text-3)', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ color: 'var(--ms-teal)', fontWeight: 700 }}>{doneCount}</span> / {items.length} done
            </div>
          </div>
          <div className="ms-checklist-list">
            {items.map((item, i) => {
              const key = `${activeWeek}-${i}`;
              return (
                <div
                  key={key}
                  className={`ms-check-item${checked[key] ? ' checked' : ''}`}
                  onClick={() => toggle(key)}
                  role="checkbox"
                  aria-checked={!!checked[key]}
                  tabIndex={0}
                  onKeyDown={e => e.key === ' ' && toggle(key)}
                >
                  <div className="ms-check-box">{checked[key] ? '✓' : ''}</div>
                  <span className="ms-check-text">{item}</span>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </section>
  );
}

/* ── Final CTA ── */
function FinalCTA() {
  return (
    <section className="ms-final-cta" id="ms-final">
      <div className="ms-final-cta-bg" />
      <div className="ms-reveal ms-final-cta-content">
        <h2 className="ms-cta-title">
          The story is ready.<br />
          <span className="ms-hero-accent">Now it's time to tell it.</span>
        </h2>
        <p className="ms-cta-sub">
          This is your 30-day plan. Every hook is written. Every format is specified. Every CTA is deliberate. Start Day 1.
        </p>
        <div className="ms-cta-row">
          <button className="ms-btn-primary" onClick={() => document.getElementById('ms-calendar')?.scrollIntoView({ behavior: 'smooth' })} id="ms-final-calendar">
            View Calendar
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M5 12h14M12 5l7 7-7 7"/></svg>
          </button>
          <Link to="/logisticshq" className="ms-btn-secondary" id="ms-final-website">View Product Website →</Link>
        </div>
      </div>
    </section>
  );
}

/* ── Footer ── */
function Footer() {
  return (
    <footer className="ms-footer">
      <span>LogisticsHQ — 30-Day Marketing & Content Launch Strategy</span>
      <span style={{ margin: '0 12px', opacity: 0.3 }}>·</span>
      <span>Build the story first. Sell LogisticsHQ second.</span>
    </footer>
  );
}

/* ════════════════════════════════════════════════════════════════
   MAIN PAGE
   ════════════════════════════════════════════════════════════════ */
export default function MarketingStrategyPage() {
  useReveal();
  return (
    <div className="ms-page">
      <Navbar />
      <main>
        <Hero />
        <Phases />
        <Calendar />
        <BrandStory />
        <Platforms />
        <ContentSeries />
        <ContentFunnel />
        <CTAStrategy />
        <PostingFrequency />
        <KPIDashboard />
        <WeeklySchedule />
        <ProductionChecklist />
        <FinalCTA />
      </main>
      <Footer />
    </div>
  );
}
