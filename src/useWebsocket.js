import { useState, useRef, useEffect } from "react";

export function useWebsocket(dispatch) {
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
