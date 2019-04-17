import React, { useState, useRef, useEffect } from "react";
import ReactDOM from "react-dom";

function App() {
  const [captureStream, setCaptureStream] = useState([]);
  const [retry, setRetry] = useState(false);
  const socket = useRef(null);

  useEffect(() => {
    if (!socket.current || retry) {
      const ws = new WebSocket("ws://localhost:3000/ws");

      ws.onopen = () => console.log("Connection established");
      ws.onmessage = message => {
        try {
          const obj = JSON.parse(message.data);
          console.log(obj);
          const xml = obj.payload.split("\r\n", 2)[1];
          setCaptureStream(z => z.concat({ ...obj, xml }));
        } catch (err) {
          console.error(err);
          ws.onclose = null;
          ws.close();
        }
      };
      ws.onerror = () => console.log("Connection Error");
      ws.onclose = () => setTimeout(() => setRetry(true), 5000);

      setRetry(false);
      socket.current = ws;
    }
  }, [retry]);

  return (
    <div>
      <pre>{captureStream.map(x => x.payload)}</pre>
    </div>
  );
}

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
