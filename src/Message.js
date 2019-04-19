import React from "react";

function Message({ data }) {
  return (
    <>
      {Object.entries(data).map(
        ([k, v]) =>
          k !== "_attributes" && (
            <div key={k}>
              <pre>{k}</pre>
              <div style={{ paddingLeft: "12px" }}>
                {v._attributes &&
                  Object.entries(v._attributes).map(([k, v]) => (
                    <div key={k}>
                      <pre>{`${k}: ${v}`}</pre>
                    </div>
                  ))}
              </div>
              <div style={{ paddingLeft: "36px" }}>
                <Message data={v} />
              </div>
            </div>
          )
      )}
    </>
  );
}

export { Message };
