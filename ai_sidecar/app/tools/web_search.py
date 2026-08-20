import os
from langchain_community.tools.tavily_search import TavilySearchResults
from langchain_community.tools import DuckDuckGoSearchRun

def get_web_search_tool():
    """
    Returns a tool for performing web searches.
    Uses Tavily if TAVILY_API_KEY is configured in env, else falls back to DuckDuckGo.
    """
    tavily_key = os.getenv("TAVILY_API_KEY")
    if tavily_key:
        print("[AI Sidecar Tools] Initializing TavilySearchResults tool.")
        return TavilySearchResults(
            max_results=3,
            search_depth="advanced",
            include_answer=True
        )
    else:
        print("[AI Sidecar Tools] TAVILY_API_KEY not found in environment. Falling back to DuckDuckGoSearchRun.")
        return DuckDuckGoSearchRun()
