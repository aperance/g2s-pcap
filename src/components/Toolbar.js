import React from "react";

function Toolbar({ dispatch }) {
  return (
    <div className="toolbar">
      <label htmlFor="id">EGM ID</label>
      <input
        type="text"
        id="id"
        name="id"
        onChange={e => dispatch({ type: "setFilter", data: e.target.value })}
      />
    </div>
  );
}

export { Toolbar };
