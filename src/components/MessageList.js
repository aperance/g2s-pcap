import React from "react";
import { AutoSizer, List } from "react-virtualized";
import { Message } from "./Message";

function MessageList({ state }) {
  return (
    <AutoSizer>
      {({ height, width }) => (
        <List
          width={width}
          height={height}
          rowCount={state.length}
          rowHeight={({ index }) => state[index].height}
          rowRenderer={({ key, index, style }) =>
            state[index].xml && (
              <div key={key} style={style}>
                <Message data={state[index].xml} />
              </div>
            )
          }
        />
      )}
    </AutoSizer>
  );
}

export { MessageList };
