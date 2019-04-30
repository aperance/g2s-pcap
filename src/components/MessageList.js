import React, { useRef, useEffect, useState } from "react";
import { AutoSizer, List } from "react-virtualized";
import { Message } from "./Message";

function MessageList({ state }) {
  const listEl = useRef(null);
  const [autoScroll, setAutoScroll] = useState(false);

  useEffect(() => {
    if (autoScroll) listEl.current.scrollToRow(state.length);
  }, [state]);

  return (
    <div className="message-list">
      <AutoSizer>
        {({ height, width }) => (
          <List
            ref={listEl}
            width={width}
            height={height}
            rowCount={state.length}
            rowHeight={({ index }) => state[index].height}
            rowRenderer={({ key, index, style }) =>
              state[index].message && (
                <div key={key} style={style}>
                  <Message state={state[index]} />
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
