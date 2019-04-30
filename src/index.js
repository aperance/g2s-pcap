import React, { useReducer, useState, useRef, useEffect } from "react";
import ReactDOM from "react-dom";
import { formatXml } from "./formatters/xml";
import { formatHex } from "./formatters/hex";
import { MessageList } from "./components/MessageList";
import { Toolbar } from "./components/Toolbar";

function App() {
  const [state, dispatch] = useReducer(reducer, []);
  useWebsocket(dispatch);
  return (
    <div className="root">
      <MessageList state={state} />
      <Toolbar />
    </div>
  );
}

function reducer(state, action) {
  switch (action.type) {
    case "push":
      try {
        let message;
        const raw = JSON.parse(action.data);
        console.log(raw);

        if (raw.protocol === "freeform") message = formatHex(raw.payload);
        else if (raw.protocol === "g2s") message = formatXml(raw.payload);
        else return [...state];

        const height = ((message.match(/\n/g) || []).length + 5) * 16;
        return [...state, { raw, message, height }];
      } catch (err) {
        console.error(err);
        return [...state];
      }
    case "clear":
      return [];
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
      ws.onmessage = e => dispatch({ type: "push", data: e.data });

      setRetry(false);
      socket.current = ws;
    }
  }, [retry]);
}

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
