import React from "react";
import { Message } from "./Message";

function MessageList({ messages }) {
  return (
    <div className="message-list">
      {messages.map((message, index) => {
        return (
          <div key={index}>
            <Message message={message.formattedMessage} />
          </div>
        );
      })}
    </div>
  );
}

export { MessageList };
