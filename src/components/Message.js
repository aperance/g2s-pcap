import React, { useRef, useLayoutEffect } from "react";
import hljs from "highlight.js/lib/highlight";
import xml from "highlight.js/lib/languages/xml";
hljs.registerLanguage("xml", xml);
import "highlight.js/styles/atom-one-dark-reasonable.css";

function Message({ data }) {
  const rootEl = useRef(null);

  useLayoutEffect(() => {
    rootEl.current.querySelectorAll("pre code").forEach(block => {
      hljs.highlightBlock(block);
    });
  }, []);

  return (
    <div ref={rootEl}>
      <pre>
        <code className="xml">{data}</code>
      </pre>
    </div>
  );
}

export { Message };
