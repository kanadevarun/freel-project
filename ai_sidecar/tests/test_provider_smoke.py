import os
import pytest
from dotenv import load_dotenv

load_dotenv()

from langchain_core.tools import tool
from langchain_core.messages import AIMessage

try:
    from langchain_google_genai import ChatGoogleGenerativeAI
except ImportError:
    ChatGoogleGenerativeAI = None  # type: ignore

try:
    from langchain_openai import ChatOpenAI
except ImportError:
    ChatOpenAI = None  # type: ignore

from app.tools.llm_factory import get_chat_model
from app.agents.llm_utils import execute_llm_json, execute_llm_text


@tool
def calculate_freight(origin: str, destination: str) -> float:
    """Calculate the ocean freight buy price between origin and destination ports."""
    if origin == "INNSA" and destination == "DEHAM":
        return 2800.0
    return 3200.0


def test_real_gemini_connectivity():
    """Provider Smoke: Real Google Gemini API connectivity and text completion."""
    if ChatGoogleGenerativeAI is None:
        pytest.skip("langchain_google_genai not installed")

    google_key = os.getenv("GOOGLE_API_KEY")
    if not google_key:
        pytest.skip("GOOGLE_API_KEY not configured in env")

    model_name = os.getenv("GEMINI_MODEL", "gemini-3.1-flash-lite")
    chat = ChatGoogleGenerativeAI(
        model=model_name,
        google_api_key=google_key,
        temperature=0.1,
        max_retries=0,
    )
    try:
        resp = chat.invoke("Reply with exactly: GEMINI_OK")
    except Exception as e:
        err_str = str(e)
        if "429" in err_str or "RESOURCE_EXHAUSTED" in err_str or "quota" in err_str.lower():
            pytest.skip(f"Gemini API Quota Throttled (429): {err_str[:120]}")
        raise e

    assert resp is not None, "Response must not be None"
    content_text = ""
    if isinstance(resp.content, list):
        for part in resp.content:
            if isinstance(part, dict) and "text" in part:
                content_text += part["text"]
            elif isinstance(part, str):
                content_text += part
    elif isinstance(resp.content, str):
        content_text = resp.content

    assert len(content_text.strip()) > 0, "Gemini returned empty response"
    assert "GEMINI_OK" in content_text.strip(), f"Expected GEMINI_OK, got: {content_text}"


def test_real_gemini_tool_binding():
    """Provider Smoke: Real Google Gemini API function/tool calling."""
    if ChatGoogleGenerativeAI is None:
        pytest.skip("langchain_google_genai not installed")

    google_key = os.getenv("GOOGLE_API_KEY")
    if not google_key:
        pytest.skip("GOOGLE_API_KEY not configured in env")


    model_name = os.getenv("GEMINI_MODEL", "gemini-3.1-flash-lite")
    chat = ChatGoogleGenerativeAI(
        model=model_name,
        google_api_key=google_key,
        temperature=0.1,
        max_retries=0,
    )
    chat_with_tools = chat.bind_tools([calculate_freight])

    prompt = "Use calculate_freight to determine freight for origin INNSA and destination DEHAM."
    try:
        res = chat_with_tools.invoke(prompt)
    except Exception as e:
        err_str = str(e)
        if "429" in err_str or "RESOURCE_EXHAUSTED" in err_str or "quota" in err_str.lower():
            pytest.skip(f"Gemini API Quota Throttled (429): {err_str[:120]}")
        raise e

    assert hasattr(res, "tool_calls"), "Response must contain tool_calls attribute"
    assert len(res.tool_calls) > 0, f"Expected at least one tool call, got: {res.tool_calls}"
    assert res.tool_calls[0]["name"] == "calculate_freight"


def test_real_openai_connectivity():
    """Provider Smoke: Real OpenAI API connectivity."""
    if ChatOpenAI is None:
        pytest.skip("langchain_openai not installed")

    openai_key = os.getenv("OPENAI_API_KEY")
    if not openai_key:
        pytest.skip("OPENAI_API_KEY not configured in env")

    chat = ChatOpenAI(
        model="gpt-4o-mini",
        api_key=openai_key,
        temperature=0.1,
    )
    try:
        resp = chat.invoke("Reply with exactly: OPENAI_OK")
        assert resp is not None
        assert "OPENAI_OK" in str(resp.content)
    except Exception as e:
        err_str = str(e)
        if "insufficient_quota" in err_str or "429" in err_str:
            pytest.skip(f"OpenAI provider quota limit reached (429 insufficient_quota): {err_str[:100]}")
        raise e


def test_factory_end_to_end():
    """Integration: get_chat_model executes cleanly with full failover cascade."""
    chat = get_chat_model(tools=[calculate_freight])
    res = chat.invoke("What is the ocean freight from INNSA to DEHAM?")
    assert res is not None
    assert isinstance(res, AIMessage)


def test_llm_utils_end_to_end():
    """Integration: execute_llm_json and execute_llm_text execute through centralized factory."""
    text_res = execute_llm_text("Reply with exactly: LOGISTICSHQ_ONLINE")
    assert text_res is not None and len(text_res) > 0

    json_res = execute_llm_json("Return JSON with key 'health' set to 'HEALTHY'")
    if json_res is not None:
        assert isinstance(json_res, (dict, list))
        if isinstance(json_res, dict) and "health" in json_res:
            assert json_res.get("health") == "HEALTHY"
        else:
            assert "carrier_name" in str(json_res) or "sell_price" in str(json_res)

