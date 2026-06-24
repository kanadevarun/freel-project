import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../../context/AuthContext';
import { saveOnboardingStep, completeOnboarding as backendCompleteOnboarding, skipOnboarding } from '../../../services/authService';
import { Check } from 'lucide-react';
import './OnboardingPage.css';

// ── ROLE-SPECIFIC QUESTIONS ──
const QUESTIONS = {
  SHIPPER: [
    {
      step: 1,
      key: 'primary_transport',
      title: 'How do you primarily move goods?',
      type: 'single',
      options: ['Air Freight', 'Ocean Freight', 'Road Freight', 'Multi-Modal'],
    },
    {
      step: 2,
      key: 'monthly_volume',
      title: 'What is your average shipment volume?',
      type: 'single',
      options: ['1–10 Shipments / Month', '11–50 Shipments / Month', '51–200 Shipments / Month', '200+ Shipments / Month'],
    },
    {
      step: 3,
      key: 'trade_regions',
      title: 'Which regions do you trade with most frequently?',
      type: 'multi',
      options: ['India', 'Middle East', 'Europe', 'North America', 'Southeast Asia', 'Australia'],
    },
    {
      step: 4,
      key: 'primary_goal',
      title: 'What are you looking for from Freel?',
      type: 'multi',
      options: ['Better Freight Rates', 'Shipment Tracking', 'Carrier Discovery', 'RFQ Management', 'Trade Intelligence'],
    },
  ],
  CARRIER: [
    {
      step: 1,
      key: 'transport_modes',
      title: 'What transport modes do you operate?',
      type: 'multi',
      options: ['Full Truckload (FTL)', 'Less Than Truckload (LTL)', 'Container Transport', 'Air Cargo', 'Ocean Freight', 'Multi-Modal'],
    },
    {
      step: 2,
      key: 'primary_regions',
      title: 'Which regions do you primarily serve?',
      type: 'multi',
      options: ['India Domestic', 'Middle East', 'Asia Pacific', 'Europe', 'North America', 'Global'],
    },
    {
      step: 3,
      key: 'fleet_size',
      title: 'How large is your fleet or logistics network?',
      type: 'single',
      options: ['1–10 Vehicles', '11–50 Vehicles', '51–200 Vehicles', '200+ Vehicles', 'Asset-Light / Partner Network'],
    },
    {
      step: 4,
      key: 'cargo_types',
      title: 'What cargo types do you handle?',
      type: 'multi',
      options: ['General Cargo', 'Perishables', 'Pharmaceuticals', 'Hazardous Goods', 'E-commerce', 'Automotive', 'Oversized Cargo'],
    },
  ],
  FREIGHT_FORWARDER: [
    {
      step: 1,
      key: 'provided_services',
      title: 'Which services do you provide?',
      type: 'multi',
      options: ['Air Freight', 'Ocean Freight', 'Customs Clearance', 'Warehousing', 'Road Transport', 'Project Cargo'],
    },
    {
      step: 2,
      key: 'monthly_shipments',
      title: 'How many shipments do you manage monthly?',
      type: 'single',
      options: ['1–50', '51–200', '201–1000', '1000+'],
    },
    {
      step: 3,
      key: 'operating_markets',
      title: 'Which markets do you operate in?',
      type: 'multi',
      options: ['India', 'Middle East', 'Europe', 'North America', 'Asia Pacific', 'Global'],
    },
    {
      step: 4,
      key: 'biggest_challenge',
      title: 'What is your biggest challenge today?',
      type: 'single',
      options: ['Finding Shippers', 'Managing RFQs', 'Shipment Visibility', 'Documentation', 'Carrier Management'],
    },
  ],
};

export default function OnboardingPage() {
  const navigate = useNavigate();
  const { org, updateOrg, completeOnboarding: contextCompleteOnboarding } = useAuth();
  
  // Default to SHIPPER if org isn't loaded yet
  const roleType = org?.orgType || 'SHIPPER';
  const questions = QUESTIONS[roleType] || QUESTIONS.SHIPPER;

  const [currentStep, setCurrentStep] = useState(1);
  const [answers, setAnswers] = useState({});
  const [isLoading, setIsLoading] = useState(false);
  const [showCreationModal, setShowCreationModal] = useState(false);
  const [loadingStep, setLoadingStep] = useState(0);

  const currentQ = questions.find(q => q.step === currentStep);
  const isLastStep = currentStep === questions.length;

  const handleToggleOption = (option) => {
    if (currentQ.type === 'single') {
      setAnswers({ ...answers, [currentQ.key]: option });
    } else {
      const currentSelection = answers[currentQ.key] || [];
      if (currentSelection.includes(option)) {
        setAnswers({ ...answers, [currentQ.key]: currentSelection.filter(item => item !== option) });
      } else {
        setAnswers({ ...answers, [currentQ.key]: [...currentSelection, option] });
      }
    }
  };

  const runCreationSequence = async () => {
    setShowCreationModal(true);
    setLoadingStep(1);
    await new Promise(r => setTimeout(r, 600));
    setLoadingStep(2);
    await new Promise(r => setTimeout(r, 600));
    setLoadingStep(3);
    await new Promise(r => setTimeout(r, 600));
    setLoadingStep(4);
    await new Promise(r => setTimeout(r, 600));
    setLoadingStep(5);
    await new Promise(r => setTimeout(r, 1200));
    
    await contextCompleteOnboarding(answers);
    navigate('/dashboard');
  };

  const handleNext = async () => {
    const currentAnswer = answers[currentQ.key];
    if (!currentAnswer || (Array.isArray(currentAnswer) && currentAnswer.length === 0)) {
      return; // Must answer to proceed
    }

    setIsLoading(true);
    try {
      // Save step to backend
      await saveOnboardingStep({
        orgId: org?.id,
        step: currentStep,
        questionKey: currentQ.key,
        answer: { value: currentAnswer }
      });

      if (isLastStep) {
        if (typeof backendCompleteOnboarding === 'function') {
           await backendCompleteOnboarding();
        }
        updateOrg({ onboardingCompleted: true });
        runCreationSequence();
      } else {
        setCurrentStep(currentStep + 1);
      }
    } catch (err) {
      console.error('Failed to save step:', err);
    } finally {
      if (!isLastStep) setIsLoading(false);
    }
  };

  const handleSkip = async () => {
    setIsLoading(true);
    try {
      await skipOnboarding();
      updateOrg({ onboardingCompleted: true });
      runCreationSequence();
    } catch (err) {
      console.error('Failed to skip:', err);
      setIsLoading(false);
    }
  };

  // Progress bar width
  const progressPercent = ((currentStep - 1) / questions.length) * 100;

  const hasAnsweredCurrent = () => {
    const ans = answers[currentQ.key];
    return ans && (!Array.isArray(ans) || ans.length > 0);
  };

  return (
    <div className="onboarding-page">
      {/* Top Progress Bar */}
      <div className="onboarding-progress-container">
        <div className="onboarding-progress-bar" style={{ width: `${progressPercent}%` }} />
      </div>

      <div className="onboarding-content animate-fade-in-up">
        <div className="onboarding-header">
          <h2>Welcome to Freel</h2>
          <p>Let's customize your {roleType.toLowerCase().replace('_', ' ')} experience. (Step {currentStep} of {questions.length})</p>
        </div>

        <div className="onboarding-question-card card-premium">
          <h3>{currentQ.title}</h3>
          
          <div className="onboarding-options-grid">
            {currentQ.options.map(option => {
              const isSelected = currentQ.type === 'single' 
                ? answers[currentQ.key] === option 
                : (answers[currentQ.key] || []).includes(option);

              return (
                <button
                  key={option}
                  className={`onboarding-option-btn ${isSelected ? 'selected' : ''}`}
                  onClick={() => handleToggleOption(option)}
                >
                  <div className={`onboarding-option-checkbox ${currentQ.type === 'single' ? 'is-radio' : ''}`}>
                    {isSelected && <Check size={14} color="white" />}
                  </div>
                  <span>{option}</span>
                </button>
              );
            })}
          </div>

          <div className="onboarding-actions-divider" />
          
          <div className="onboarding-actions">
            <button className="onboarding-skip-btn" onClick={handleSkip} disabled={isLoading}>
              Skip for now
            </button>
            
            <button 
              className="onboarding-next-btn btn-primary" 
              onClick={handleNext} 
              disabled={!hasAnsweredCurrent() || isLoading}
            >
              {isLoading ? <div className="auth-spinner" /> : (isLastStep ? 'Complete Setup →' : 'Next Step →')}
            </button>
          </div>
        </div>
      </div>
      {showCreationModal && (
        <div className="onboarding-modal-overlay">
          <div className="onboarding-creation-modal">
            {loadingStep < 5 ? (
              <>
                <div className="modal-spinner-large" />
                <h3>Setting Up Your Freight Workspace</h3>
                <p className="modal-subtitle">We're preparing your organization and configuring your logistics environment.</p>
                <ul className="creation-steps-list">
                  <li className={loadingStep >= 1 ? 'done' : ''}>
                    <div className="status-icon">{loadingStep >= 1 ? <Check size={14}/> : <div className="spinner-small"/>}</div>
                    Creating Organization
                  </li>
                  <li className={loadingStep >= 2 ? 'done' : ''}>
                    <div className="status-icon">{loadingStep >= 2 ? <Check size={14}/> : (loadingStep > 1 ? <div className="spinner-small"/> : <div className="dot"/>)}</div>
                    Assigning Account Owner
                  </li>
                  <li className={loadingStep >= 3 ? 'done' : ''}>
                    <div className="status-icon">{loadingStep >= 3 ? <Check size={14}/> : (loadingStep > 2 ? <div className="spinner-small"/> : <div className="dot"/>)}</div>
                    Configuring Logistics Profile
                  </li>
                  <li className={loadingStep >= 4 ? 'done' : ''}>
                    <div className="status-icon">{loadingStep >= 4 ? <Check size={14}/> : (loadingStep > 3 ? <div className="spinner-small"/> : <div className="dot"/>)}</div>
                    Enabling Trade Intelligence
                  </li>
                  <li className={loadingStep >= 5 ? 'done' : ''}>
                    <div className="status-icon">{loadingStep >= 5 ? <Check size={14}/> : (loadingStep > 4 ? <div className="spinner-small"/> : <div className="dot"/>)}</div>
                    Activating Account
                  </li>
                </ul>
              </>
            ) : (
              <div className="creation-success-state animate-pop-in">
                <div className="success-icon-large">
                  <Check size={40} color="white" />
                </div>
                <h3>Organization Ready</h3>
                <p>Welcome to Freel.<br/>Your logistics workspace is now active.</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
