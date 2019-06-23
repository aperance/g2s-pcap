import React, { useState, useRef, useEffect, useLayoutEffect } from "react";
import hljs from "highlight.js/lib/highlight";
import plaintext from "highlight.js/lib/languages/plaintext";
import xml from "highlight.js/lib/languages/xml";
hljs.registerLanguage("plaintext", plaintext);
hljs.registerLanguage("xml", xml);
import "highlight.js/styles/atom-one-dark-reasonable.css";

function Message({ message, filters }) {
  const [visible, setVisible] = useState(true);
  const rootEl = useRef(null);

  useLayoutEffect(() => {
    rootEl.current.querySelectorAll("pre code").forEach(block => {
      hljs.highlightBlock(block);
    });
  }, [message, visible]);

  useEffect(() => {
    try {
      const pattern =
        message.raw.protocol === "G2S"
          ? `(?<=egmId=").*?(?=")`
          : `(?<=AssetID:\\s\\w+\\s\\().*?(?=\\))`;
      const regex = RegExp(pattern, "g");
      setVisible(
        message.formattedMessage.match(regex)[0].includes(filters.egmId)
      );
    } catch (e) {
      console.error(e);
      setVisible(true);
    }
  }, [message.formattedMessage, filters]);

  return (
    <div ref={rootEl}>
      {visible && (
        <pre>
          <code className="xml" style={{ overflowX: "hidden" }}>
            {message.formattedMessage}
          </code>
        </pre>
      )}
    </div>
  );
}

export { Message };
