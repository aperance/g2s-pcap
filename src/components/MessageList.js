import React from "react";
import { AutoSizer, List } from "react-virtualized";
import { Message } from "./Message";

function MessageList({ state }) {
  const rowRenderer = ({ key, index, style }) =>
    state[index].obj && (
      <div key={key} style={style}>
        <Message data={state[index].obj} />
      </div>
    );

  const getRowHeight = ({ index }) =>
    state[index].obj
      ? 20 * JSON.stringify(state[index].obj, null, 2).match(/\n/g).length
      : 0;

  return (
    <AutoSizer>
      {({ height, width }) => (
        <List
          width={width}
          height={height}
          rowCount={state.length}
          rowHeight={getRowHeight}
          rowRenderer={rowRenderer}
        />
      )}
    </AutoSizer>
  );
}

export { MessageList };
