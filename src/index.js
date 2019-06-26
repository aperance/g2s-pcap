import React, {
  useReducer,
  useState,
  useRef,
  useEffect,
  useLayoutEffect
} from "react";
import ReactDOM from "react-dom";
import { formatXml } from "./formatters/xml";
import { formatHex } from "./formatters/hex";
import { Message } from "./components/Message";
import { Toolbar } from "./components/Toolbar";

function App() {
  const [state, dispatch] = useReducer(reducer, {
    messages: [],
    filters: { egmId: "", general: "" }
  });
  const ref = useAutoScroll(state);
  useWebsocket(dispatch);

  return (
    <div className="root" ref={ref}>
      {state.messages.map((message, index) => (
        <div key={index}>
          <Message message={message} filters={state.filters} />
        </div>
      ))}
      <Toolbar dispatch={dispatch} filters={state.filters} />
    </div>
  );
}

function reducer(state, action) {
  switch (action.type) {
    case "pushMessage":
      try {
        const { protocol, payload, flow, direction } = action.data;
        let formattedMessage = `${direction} ${protocol} Message (${flow})\n\n`;

        if (protocol === "Freeform") formattedMessage += formatHex(payload);
        else if (protocol === "G2S") formattedMessage += formatXml(payload);
        else return { ...state };

        const height = ((formattedMessage.match(/\n/g) || []).length + 3) * 16;

        return {
          ...state,
          messages: [
            ...state.messages,
            { formattedMessage, height, raw: action.data }
          ]
        };
      } catch (err) {
        console.error(err);
        return { ...state };
      }
    case "setFilter":
      return {
        ...state,
        filters: { ...state.filters, [action.name]: action.data }
      };
    case "clearState":
      return { ...state, messages: [] };
    default:
      throw new Error();
  }
}

function useAutoScroll({ messages }) {
  const ref = useRef(null);
  const [autoScroll, setAutoScroll] = useState(true);

  useEffect(() => {
    ref.current.addEventListener("scroll", e => {
      const threshold = e.target.scrollHeight - e.target.clientHeight - 100;
      setAutoScroll(e.target.scrollTop > threshold);
    });
  }, []);

  useLayoutEffect(() => {
    if (!autoScroll) return;
    ref.current.scrollTop = ref.current.scrollHeight - ref.current.clientHeight;
  }, [messages]);

  return ref;
}

function useWebsocket(dispatch) {
  const [retry, setRetry] = useState(false);
  const socket = useRef(null);

  useEffect(() => {
    if (!socket.current || retry) {
      const ws = new WebSocket("ws://10.91.1.46:3000/ws");

      ws.onopen = () => console.log("Connection established");
      ws.onerror = () => {
        console.error("Connection Error");
        dispatch({ type: "clearState" });
      };
      ws.onclose = () => setTimeout(() => setRetry(true), 5000);
      ws.onmessage = e => {
        const message = JSON.parse(e.data);
        console.log(message);
        dispatch({ type: "pushMessage", data: message });
      };

      setRetry(false);
      socket.current = ws;
    }
  }, [retry]);
}

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
