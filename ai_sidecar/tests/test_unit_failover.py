import pytest
from langchain_core.messages import AIMessage, HumanMessage
from langchain_core.tools import tool
from app.tools.llm_factory import FailoverChatModel


class MockModelSuccess:
    def __init__(self, response_text="Success Response"):
        self.response_text = response_text
        self.tools = []

    def invoke(self, input, **kwargs):
        return AIMessage(content=self.response_text)

    def bind_tools(self, tools, **kwargs):
        new_model = MockModelSuccess(self.response_text)
        new_model.tools = tools
        return new_model


class MockModelFailure:
    def __init__(self, error_message="Provider Error 500"):
        self.error_message = error_message
        self.tools = []

    def invoke(self, input, **kwargs):
        raise RuntimeError(self.error_message)

    def bind_tools(self, tools, **kwargs):
        new_model = MockModelFailure(self.error_message)
        new_model.tools = tools
        return new_model


class MockToolCallingModel:
    def __init__(self, tool_name="calculate_freight", tool_args=None):
        self.tool_name = tool_name
        self.tool_args = tool_args or {"origin": "INNSA", "destination": "DEHAM"}
        self.tools = []

    def invoke(self, input, **kwargs):
        msg = AIMessage(content="")
        msg.tool_calls = [{"name": self.tool_name, "args": self.tool_args, "id": "call_123", "type": "tool_call"}]
        return msg

    def bind_tools(self, tools, **kwargs):
        new_model = MockToolCallingModel(self.tool_name, self.tool_args)
        new_model.tools = tools
        return new_model


@tool
def dummy_tool(arg: str) -> str:
    """Dummy tool for testing."""
    return f"Result for {arg}"


def test_unit_primary_success():
    """Unit test: Primary succeeds on first attempt without calling fallback."""
    primary = MockModelSuccess("Primary Model Output")
    fallback = MockModelFailure("Fallback should not be called")
    model = FailoverChatModel(primary=primary, fallback=fallback)

    res = model.invoke("Hello")
    assert isinstance(res, AIMessage)
    assert res.content == "Primary Model Output"


def test_unit_failover_to_fallback():
    """Unit test: Primary fails, delegates execution to secondary/fallback model."""
    primary = MockModelFailure("Primary 429 Rate Limit")
    fallback = MockModelSuccess("Secondary Model Output")
    model = FailoverChatModel(primary=primary, fallback=fallback)

    res = model.invoke("Hello")
    assert isinstance(res, AIMessage)
    assert res.content == "Secondary Model Output"


def test_unit_failover_to_deterministic_mock():
    """Unit test: Both primary and fallback fail, returns deterministic quote AIMessage."""
    primary = MockModelFailure("Primary 429")
    fallback = MockModelFailure("Fallback 429 insufficient_quota")
    model = FailoverChatModel(primary=primary, fallback=fallback)

    input_msgs = [HumanMessage(content="Origin: INNSA\nDestination: DEHAM\nIncoterms: FOB")]
    res = model.invoke(input_msgs)
    assert isinstance(res, AIMessage)
    assert "carrier_name" in res.content
    assert "sell_price" in res.content
    assert "Maersk" in res.content


def test_unit_bind_tools_preserves_failover():
    """Unit test: bind_tools returns a FailoverChatModel that preserves failover during tool calling."""
    primary_failing = MockModelFailure("Gemini Tool Binding Failed")
    fallback_working = MockToolCallingModel("dummy_tool", {"arg": "test_val"})

    model = FailoverChatModel(primary=primary_failing, fallback=fallback_working)
    bound_model = model.bind_tools([dummy_tool])

    assert isinstance(bound_model, FailoverChatModel)
    assert len(bound_model.primary.tools) == 1
    assert len(bound_model.fallback.tools) == 1

    # When invoking the bound model, primary fails and fallback returns tool calls
    res = bound_model.invoke("Use dummy tool")
    assert hasattr(res, "tool_calls")
    assert len(res.tool_calls) == 1
    assert res.tool_calls[0]["name"] == "dummy_tool"


def test_unit_mock_pricing_extraction_lane():
    """Unit test: Deterministic fallback parses lane info from prompt."""
    model = FailoverChatModel(primary=MockModelFailure(), fallback=MockModelFailure())
    
    # Custom lane test (non INNSA-DEHAM uses default 20% markup)
    input_msgs = [HumanMessage(content="Origin: NLRTM\nDestination: USNYC\nIncoterms: CIF")]
    res = model.invoke(input_msgs)
    assert "NLRTM" in res.content
    assert "USNYC" in res.content
    assert "3200.0" in res.content
