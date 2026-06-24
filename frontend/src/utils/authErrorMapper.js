export function mapAuthError(rawMessage, flow = 'default') {
  if (!rawMessage) {
    return {
      code: 'UNKNOWN_ERROR',
      message: 'An unexpected error occurred. Please try again.',
      rawMessage: ''
    };
  }

  const rawStr = String(rawMessage);

  // Common mapping definitions
  const errorMap = {
    UsernameExistsException: {
      message: 'This email is already registered. Please sign in instead.',
      action: 'LOGIN'
    },
    InvalidPasswordException: {
      message: flow === 'reset'
        ? "Your new password doesn't meet the required security criteria."
        : "Your password doesn't meet the required security criteria. Please use at least 8 characters, 1 uppercase letter, 1 number, and 1 special character.",
    },
    InvalidParameterException: {
      message: 'Please check the information you entered and try again.',
    },
    CodeMismatchException: {
      message: flow === 'reset' 
        ? 'The reset code is incorrect. Please check it and try again.'
        : 'The verification code you entered is incorrect. Please check the code and try again.',
    },
    ExpiredCodeException: {
      message: flow === 'reset'
        ? 'This reset code has expired. Please request a new one and try again.'
        : 'This verification code has expired. Please request a new code and try again.',
    },
    UserNotFoundException: {
      message: flow === 'verify' 
        ? 'We could not find an account for this email. Please sign up again.'
        : flow === 'login'
        ? 'Incorrect email or password. Please try again.' // Prevent leakage
        : flow === 'forgot'
        ? 'No account found with this email.'
        : 'User account could not be found.',
      action: flow === 'verify' ? 'SIGNUP' : null
    },
    NotAuthorizedException: {
      message: 'Incorrect email or password. Please try again.',
    },
    UserNotConfirmedException: {
      message: 'Your email address is not verified yet. Please verify your email before signing in.',
      action: 'VERIFY'
    }
  };

  // Find the matching exception code in the raw message
  let matchedKey = null;
  for (const key of Object.keys(errorMap)) {
    if (rawStr.includes(key)) {
      matchedKey = key;
      break;
    }
  }

  // Handle some specific string matches if Cognito code is missing but text is present
  if (!matchedKey && rawStr.toLowerCase().includes('already exists')) {
    matchedKey = 'UsernameExistsException';
  }

  if (matchedKey) {
    return {
      code: matchedKey,
      message: errorMap[matchedKey].message,
      action: errorMap[matchedKey].action || null,
      rawMessage: rawStr
    };
  }

  // Handle generic fallbacks by flow
  if (flow === 'signup') {
    return {
      code: 'GENERIC_SIGNUP_FAILURE',
      message: 'We couldn’t create your account right now. Please try again in a moment.',
      rawMessage: rawStr
    };
  }

  if (flow === 'forgot') {
    return {
      code: 'GENERIC_FORGOT_FAILURE',
      message: 'We couldn’t start the password reset process right now. Please try again.',
      rawMessage: rawStr
    };
  }

  if (flow === 'reset') {
    return {
      code: 'GENERIC_RESET_FAILURE',
      message: "We couldn't reset your password. Please try again.",
      rawMessage: rawStr
    };
  }

  // Generic default
  return {
    code: 'AUTH_ERROR',
    message: 'An unexpected error occurred. Please try again.',
    rawMessage: rawStr
  };
}
