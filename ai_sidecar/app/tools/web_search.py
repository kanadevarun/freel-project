import os
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

from langchain_community.tools.tavily_search import TavilySearchResults
from langchain_community.tools import DuckDuckGoSearchRun

def get_web_search_tool():
    """
    Returns a tool for performing web searches.
    Tavily is strictly the primary priority for all market intelligence and web search,
    regardless of production readiness or evaluation mode.
    Uses Tavily whenever TAVILY_API_KEY is configured.
    """
    # 1. Primary Priority: Tavily Search
    tavily_key = os.getenv("TAVILY_API_KEY")
    if tavily_key:
        try:
            print("[AI Sidecar Tools] Prioritizing TavilySearchResults as primary web search tool.")
            return TavilySearchResults(
                max_results=3,
                search_depth="advanced",
                include_answer=True,
                tavily_api_key=tavily_key,
                handle_tool_error=True
            )
        except Exception as e:
            print(f"[AI Sidecar Tools] Failed to initialize Tavily search: {e}. Attempting fallback...")

    # 2. Secondary Fallback: Deterministic Test Mode Safeguard (only if Tavily is unavailable)
    try:
        from app.eval.harness import DeterministicTestHarness
        if DeterministicTestHarness.get_instance().is_enabled() or os.getenv("APP_ENV") == "test":
            from langchain_core.tools import tool

            @tool
            def deterministic_web_search(query: str) -> str:
                """Deterministic mock search tool returning stable market rate intelligence."""
                return f"[DETERMINISTIC_SEARCH_MOCK] Market intel for '{query}': No abnormal GRI or port congestion detected. Peak season surcharges stable at $150/TEU."

            return deterministic_web_search
    except ImportError:
        pass

    # 3. Tertiary Fallback: DuckDuckGo
    print("[AI Sidecar Tools] TAVILY_API_KEY not found in environment. Falling back to DuckDuckGoSearchRun.")
    return DuckDuckGoSearchRun()
