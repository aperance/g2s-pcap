import React, { useState, useRef, useEffect } from "react";
import ReactDOM from "react-dom";

function App() {
  const [status, setStatus] = useState("disconnected");
  const [retry, setRetry] = useState(false);
  const socket = useRef(null);

  useEffect(() => {
    if (!socket.current || retry) {
      const ws = new WebSocket("ws://localhost:3000/ws");

      ws.onopen = () => setStatus("connected");
      ws.onmessage = x => {
        try {
          console.log(JSON.parse(x.data));
        } catch (err) {
          console.error(err);
          ws.onclose = null;
          ws.close();
          setStatus("dataError");
        }
      };
      ws.onerror = () => setStatus("connectionError");
      ws.onclose = () => setTimeout(() => setRetry(true), 5000);

      setRetry(false);
      socket.current = ws;
    }
  }, [retry]);

  return <div>{status}</div>;
}

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
