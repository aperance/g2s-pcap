import React, { useReducer } from "react";
import ReactDOM from "react-dom";
import { xml2js } from "xml-js";
import { AutoSizer, List } from "react-virtualized";
import { useWebsocket } from "./useWebsocket";

function App() {
  const [state, dispatch] = useReducer(reducer, []);
  useWebsocket(dispatch);

  const rowRenderer = ({ key, index, style }) => {
    return (
      <div key={key} style={style}>
        {state[index].obj && (
          <pre>{JSON.stringify(state[index].obj, null, 2)}</pre>
        )}
      </div>
    );
  };

  const getRowHeight = ({ index }) =>
    state[index].obj
      ? 15 * (2 + JSON.stringify(state[index].obj, null, 2).match(/\n/g).length)
      : 0;

  return (
    <div style={{ height: "90vh", width: "90vw" }}>
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
    </div>
  );
}

function reducer(state, action) {
  switch (action.type) {
    case "push":
      try {
        const message = JSON.parse(action.data);
        console.log(message);

        const xml = message.payload
          .replace(/&lt;/g, "<")
          .replace(/&gt;/g, ">")
          .split(/\s*<\/*(g2s:)*g2sMessage.*?>\s*/g)[2];
        if (xml === undefined) return [...state];
        console.log(xml);

        const obj = xml2js(xml, { compact: true });
        console.log(obj);

        return [...state, { ...message, xml, obj }];
      } catch (err) {
        console.error(err);
        return [...state];
      }
    case "clear":
      return [];
    default:
      throw new Error();
  }
}

var mountNode = document.getElementById("app");
ReactDOM.render(<App />, mountNode);
