export function formatHex(raw) {
  const parsedMessage = [...parseHeader(raw), ...parseData(raw)];
  const formattedMessage = parsedMessage.reduce(
    (acc, [key, value], i, array) => {
      let result = acc + `\t${key}: ${value}`;
      if (i < array.length - 1) result += "\n";
      return result;
    },
    ""
  );
  return formattedMessage;
}

function parseHeader(raw) {
  const lookupTable = [
    ["MessageType", 1, null],
    [
      "NetworkAddress",
      16,
      x =>
        `${hexToInt(x.slice(24, 26))}.` +
        `${hexToInt(x.slice(26, 28))}.` +
        `${hexToInt(x.slice(28, 30))}.` +
        `${hexToInt(x.slice(30, 32))}`
    ],
    ["AssetID", 4, x => hexToInt(x)],
    ["CompressionType", 1, null],
    ["EncryptionType", 1, null],
    ["CommandID", 1, null],
    ["SessionID", 1, null],
    ["TransactionID", 4, null]
  ];

  const start = 0;
  const end = 58;
  const header = raw.slice(start, end);

  let index = 0;
  return lookupTable.map(([key, length, transform]) => {
    const start = index;
    const end = index + length * 2;
    let result = header.slice(start, end);
    if (transform !== null) result = `${result} (${transform(result)})`;
    index = end;
    return [key, result];
  });
}

function parseData(raw) {
  const start = 62;
  const end = start + hexToInt(raw.slice(58, 62)) * 2;
  const data = raw.slice(start, end);

  if (data.length === 0) return [];
  const paramStart = 4;
  const paramEnd = 4 + hexToInt(data.slice(2, 4)) * 2;
  return [
    [data.slice(0, 2), data.slice(paramStart, paramEnd)],
    ...parseData(data.slice(paramEnd, data.length))
  ];
}

function hexToInt(hex) {
  return parseInt(`0x${hex}`);
}
