import { useSelector } from "@legendapp/state/react";
import { memo, useRef, useState } from "react";
import { CONFIG } from "../config";
import { deleteEntry, upsertEntry } from "./actions";
import { store$ } from "./store";

interface RowProps {
  k: string;
  top: number;
}

export const Row = memo(function Row({ k, top }: RowProps) {

  const renders = useRef(0);
  renders.current += 1;

  const entry = useSelector(() => store$.entries[k].get());

  const [draft, setDraft] = useState<string | null>(null);

  if (!entry) return null;

  const save = () => {
    if (draft !== null && draft !== entry.value) void upsertEntry(k, draft);
    setDraft(null);
  };

  return (
    <div className="row" style={{ top, height: CONFIG.rowHeight }}>
      <span className="row-key" title={k}>{k}</span>

      {draft === null ? (
        <span className="row-value" onDoubleClick={() => setDraft(entry.value)} title="double-clic pour éditer">
          {entry.value}
        </span>
      ) : (
        <input
          className="row-edit"
          value={draft}
          autoFocus
          onChange={(e) => setDraft(e.target.value)}
          onBlur={save}
          onKeyDown={(e) => {
            if (e.key === "Enter") save();
            if (e.key === "Escape") setDraft(null);
          }}
        />
      )}

      {entry.expiresAt !== null && (
        <span className="row-ttl" title={`expire à ${new Date(entry.expiresAt).toLocaleTimeString("fr-FR")}`}>
          ⏳
        </span>
      )}

      <span className="row-renders" title="nombre de renders de cette ligne">
        {renders.current}
      </span>

      <button className="row-delete" onClick={() => void deleteEntry(k)} title={`DELETE ${k}`}>
        ✕
      </button>
    </div>
  );
});
