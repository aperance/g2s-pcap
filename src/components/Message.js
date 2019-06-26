import React, { useState, useRef, useEffect, useLayoutEffect } from "react";
import hljs from "highlight.js/lib/highlight";
import xml from "highlight.js/lib/languages/xml";
hljs.registerLanguage("xml", xml);
import "highlight.js/styles/atom-one-dark-reasonable.css";

function Message({ message, filters }) {
  const [visible, setVisible] = useState(false);
  const rootEl = useRef(null);

  useEffect(() => {
    setVisible(
      message.raw.egmId.includes(filters.egmId) &&
        message.formattedMessage.includes(filters.general)
    );
  }, [message.raw.egmId, filters.egmId, filters.general]);

  useLayoutEffect(() => {
    if (rootEl.current) hljs.highlightBlock(rootEl.current);
  }, [message, visible]);

  return (
    visible && (
      <pre>
        <code ref={rootEl} className="xml">
          {message.formattedMessage}
        </code>
      </pre>
    )
  );
}

export { Message };
