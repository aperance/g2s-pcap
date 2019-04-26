import React, { useRef, useLayoutEffect } from "react";
import hljs from "highlight.js/lib/highlight";
import plaintext from "highlight.js/lib/languages/plaintext";
import xml from "highlight.js/lib/languages/xml";
hljs.registerLanguage("plaintext", plaintext);
hljs.registerLanguage("xml", xml);
import "highlight.js/styles/atom-one-dark-reasonable.css";

function Message({ state: { raw, message } }) {
  const rootEl = useRef(null);

  useLayoutEffect(() => {
    rootEl.current.querySelectorAll("pre code").forEach(block => {
      hljs.highlightBlock(block);
    });
  }, []);

  return (
    <div ref={rootEl}>
      <pre>
        <code className="xml">{`  ${raw.ip} (${raw.port})\n\n${message}`}</code>
      </pre>
    </div>
  );
}

export { Message };
