import { useSelector } from "@legendapp/state/react";
import { useState } from "react";
import { CONFIG } from "../config";
import { Row } from "./Row";
import { store$ } from "./store";

export function VirtualList() {
  const keys = useSelector(() => store$.keys.get());
  const [scrollTop, setScrollTop] = useState(0);

  const { rowHeight, viewportHeight, overscan } = CONFIG;
  const startIndex = Math.max(0, Math.floor(scrollTop / rowHeight) - overscan);
  const endIndex = Math.min(
    keys.length,
    Math.ceil((scrollTop + viewportHeight) / rowHeight) + overscan,
  );
  const visible = keys.slice(startIndex, endIndex);

  return (
    <div>
      <div className="list-meta">
        {keys.length.toLocaleString("fr-FR")} entrées — {visible.length} lignes montées dans le DOM
        (n° {startIndex + 1} à {endIndex})
      </div>

      <div
        className="viewport"
        style={{ height: viewportHeight }}
        onScroll={(e) => setScrollTop(e.currentTarget.scrollTop)}
      >
        {}
        <div className="spacer" style={{ height: keys.length * rowHeight }}>
          {visible.map((key, i) => (

            <Row key={key} k={key} top={(startIndex + i) * rowHeight} />
          ))}
        </div>
      </div>
    </div>
  );
}
