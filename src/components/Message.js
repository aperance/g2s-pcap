import React from "react";

function Message({ data }) {
  return Object.entries(data).map(
    ([key, value]) =>
      key !== "_attributes" && (
        <div key={key}>
          <pre>{key}</pre>
          <div style={{ paddingLeft: "12px" }}>
            {value._attributes &&
              Object.entries(value._attributes).map(([key, value]) => (
                <div key={key}>
                  <pre>{`${key}: ${value}`}</pre>
                </div>
              ))}
          </div>
          <div style={{ paddingLeft: "36px" }}>
            <Message data={value} />
          </div>
        </div>
      )
  );
}

export { Message };
