import React, { useState, useRef, useEffect } from "react";
import ReactDOM from "react-dom";
import { parseString } from "xml2js";

function App() {
  const [captureStream, setCaptureStream] = useState([]);
  const [retry, setRetry] = useState(false);
  const socket = useRef(null);

  useEffect(() => {
    if (!socket.current || retry) {
      const ws = new WebSocket("ws://10.91.1.46:3000/ws");

      ws.onopen = () => console.log("Connection established");
      ws.onmessage = message => {
        try {
          const obj = JSON.parse(message.data);

          const xmlString = obj.payload
            .replace(/&lt;/g, "<")
            .replace(/&gt;/g, ">")
            .replace(/\t/g, "  ")
            // .split(/\s*<\/*s(oap-envelope)*:Body.*?>\s*/g)[2]
            .split(/\s*<\/*(g2s:)*g2sMessage.*?>\s*/g)[2];
          if (xmlString === undefined) return;

          parseString(xmlString, { explicitArray: false }, (err, xmlObj) => {
            const newObj = { ...obj, xmlString, xmlObj };
            console.log(newObj);
            setCaptureStream(x => x.concat(newObj));
          });
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
      <pre>
        {/* {captureStream.map(x => x.xmlString + "\n\n")} */}
        {captureStream.map(x => JSON.stringify(x.xmlObj, null, 2) + "\n\n")}
      </pre>
    </div>
  );
}

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
