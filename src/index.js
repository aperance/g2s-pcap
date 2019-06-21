import React, { useReducer, useState, useRef, useEffect } from "react";
import ReactDOM from "react-dom";
import { formatXml } from "./formatters/xml";
import { formatHex } from "./formatters/hex";
import { MessageList } from "./components/MessageList";
import { Toolbar } from "./components/Toolbar";

function App() {
  const [state, dispatch] = useReducer(reducer, {
    messages: [],
    filters: { egmId: "" }
  });
  useWebsocket(dispatch);

  const filterFunction = msg =>
    msg.formattedMessage
      .match(/(?<=egmId=").*?(?=")/g)[0]
      .includes(state.filters.egmId);

  return (
    <div className="root">
      <MessageList messages={state.messages.filter(filterFunction)} />
      <Toolbar dispatch={dispatch} filters={state.filters} />
    </div>
  );
}

function reducer(state, action) {
  switch (action.type) {
    case "pushMessage":
      try {
        const raw = JSON.parse(action.data);
        const { protocol, payload, flow, direction } = raw;
        console.log(raw);

        let formattedMessage = `${direction} ${protocol} Message (${flow})\n\n`;

        if (protocol === "Freeform") formattedMessage += formatHex(payload);
        else if (protocol === "G2S") formattedMessage += formatXml(payload);
        else return { ...state };

        const height = ((formattedMessage.match(/\n/g) || []).length + 3) * 16;

        return {
          ...state,
          messages: [...state.messages, { formattedMessage, height, raw }]
        };
      } catch (err) {
        console.error(err);
        return { ...state };
      }
    case "setFilter":
      return { ...state, filters: { egmId: action.data } };
    case "clearState":
      return {};
    default:
      throw new Error();
  }
}

function useWebsocket(dispatch) {
  const [retry, setRetry] = useState(false);
  const socket = useRef(null);

  useEffect(() => {
    if (!socket.current || retry) {
      const ws = new WebSocket("ws://10.91.1.46:3000/ws");

      ws.onopen = () => console.log("Connection established");
      ws.onerror = () => console.log("Connection Error");
      ws.onclose = () => setTimeout(() => setRetry(true), 5000);
      ws.onmessage = e => dispatch({ type: "pushMessage", data: e.data });

      setRetry(false);
      socket.current = ws;
    }
  }, [retry]);
}

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
