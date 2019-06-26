import React from "react";

function Toolbar({ dispatch }) {
  return (
    <div className="toolbar">
      <label htmlFor="id">EGM ID</label>
      <input
        type="text"
        id="id"
        name="id"
        onChange={e =>
          dispatch({ type: "setFilter", name: "egmId", data: e.target.value })
        }
      />
      <label htmlFor="search">SEARCH</label>
      <input
        type="text"
        id="search"
        name="search"
        onChange={e =>
          dispatch({ type: "setFilter", name: "general", data: e.target.value })
        }
      />
    </div>
  );
}

export { Toolbar };
