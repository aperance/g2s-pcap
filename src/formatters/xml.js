export function formatXml(raw) {
  const indent = depth => "\n" + "  ".repeat(depth);

  let ar = raw
      .replace(/&#xA;/g, "")
      .replace(/&lt;\?xml version="1\.0"\?&gt;/g, "")
      .replace(/&lt;/g, "<")
      .replace(/&gt;/g, ">")
      .replace(/\sxmlns.*?(?=[\s,>])/g, "")
      .replace(/>\s{0,}</g, "><")
      .replace(/</g, "~::~<")
      .replace(/"[\s]*\/>(?!\S*?Meter)/g, '"~::~/>')
      .replace(/\s(?=\S*?=".*?")(?!\S*?meterName)(?!\S*?meterValue)/g, "~::~")
      .split("~::~"),
    deep = 3,
    str = "",
    ix = 0;

  for (ix = 0; ix < ar.length; ix++) {
    // <elm></elm> //
    if (
      /^<\w/.exec(ar[ix - 1]) &&
      /^<\/\w/.exec(ar[ix]) &&
      /^<[\w:\-\.\,]+/.exec(ar[ix - 1]) ==
        /^<\/[\w:\-\.\,]+/.exec(ar[ix])[0].replace("/", "")
    )
      deep--;
    // <elm> //
    else if (
      ar[ix].search(/<\w/) > -1 &&
      ar[ix].search(/<\//) == -1 &&
      ar[ix].search(/\/>/) == -1
    )
      str += indent(deep++);
    // <elm>...</elm> //
    else if (ar[ix].search(/<\w/) > -1 && ar[ix].search(/<\//) > -1)
      str += indent(deep);
    // </elm> //
    else if (ar[ix].search(/<\//) > -1) str += indent(--deep);
    // <elm/> //
    else if (ar[ix].search(/<.*\/>/) > -1) str += indent(deep);
    // /> //
    else if (ar[ix].search(/\/>/) > -1) str += indent(--deep);
    // <? xml ... ?> //
    else if (ar[ix].search(/<\?/) > -1) str += indent(deep);
    // attr="" //
    else if (ar[ix].search(/\w/) > -1) str += indent(deep);

    str += ar[ix];
  }

  return str[0] == "\n" ? str.slice(1) : str;
}
