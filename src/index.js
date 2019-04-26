import React, { useReducer, useState, useRef, useEffect } from "react";
import ReactDOM from "react-dom";
import { xmlFormat } from "./xmlFormat";
import { MessageList } from "./components/MessageList";

function App() {
  const [state, dispatch] = useReducer(reducer, []);
  useWebsocket(dispatch);

  return <MessageList state={state} />;
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

function reducer(state, action) {
  switch (action.type) {
    case "push":
      try {
        const message = JSON.parse(action.data);
        console.log(message);

        if (message.protocol === "g2s") {
          const xml = xmlFormat(message.payload);
          const height = ((xml.match(/\n/g) || []).length + 5) * 16;

          return [...state, { ...message, g2s, height }];
        }
        if (message.protocol === "freeform") {
          console.log(message.payload);
        }
        return [...state];
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

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
