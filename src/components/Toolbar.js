import React from "react";

function Toolbar() {
  return (
    <div className="toolbar">
      <label htmlFor="ip">IP Address</label>
      <input type="text" id="ip" name="user_name" />
    </div>
  );
}

export { Toolbar };
