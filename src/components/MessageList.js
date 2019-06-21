import React, { useRef, useEffect, useState } from "react";
import { AutoSizer, List } from "react-virtualized";
import { Message } from "./Message";

function MessageList({ messages }) {
  const listEl = useRef(null);
  const [autoScroll, setAutoScroll] = useState(false);

  useEffect(() => {
    if (autoScroll) listEl.current.scrollToRow(messages.length);
  }, [messages]);

  return (
    <div className="message-list">
      <AutoSizer>
        {({ height, width }) => (
          <List
            ref={listEl}
            width={width}
            height={height}
            rowCount={messages.length}
            rowHeight={({ index }) => messages[index].height}
            rowRenderer={({ key, index, style }) =>
              messages[index].formattedMessage && (
                <div key={key} style={style}>
                  <Message message={messages[index].formattedMessage} />
                </div>
              )
            }
            onScroll={({ clientHeight, scrollHeight, scrollTop }) => {
              const diff = scrollHeight - clientHeight - scrollTop;
              diff < 100 ? setAutoScroll(true) : setAutoScroll(false);
            }}
          />
        )}
      </AutoSizer>
    </div>
  );
}

export { MessageList };
