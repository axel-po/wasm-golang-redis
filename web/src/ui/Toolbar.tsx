import { useSelector } from "@legendapp/state/react";
import { useState } from "react";
import { CONFIG } from "../config";
import type { FilterOp } from "../sdk";
import { flushNow, runQuery, seed, snapshotNow, touchRandomEntry, upsertEntry } from "./actions";
import { store$ } from "./store";

export function Toolbar() {
  const busy = useSelector(() => store$.busy.get());
  const seedProgress = useSelector(() => store$.seedProgress.get());
  const disabled = busy !== null;

  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [ex, setEx] = useState("");

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    const ttl = ex.trim() === "" ? undefined : Number(ex);
    void upsertEntry(key.trim(), value, ttl).then(() => {
      setKey("");
      setValue("");
      setEx("");
    });
  };

  return (
    <div className="toolbar">
      <form className="toolbar-group" onSubmit={submit}>
        <input name="key" placeholder="clé" value={key} onChange={(e) => setKey(e.target.value)} />
        <input name="value" placeholder="valeur" value={value} onChange={(e) => setValue(e.target.value)} />
        <input
          name="ex"
          placeholder="EX (s)"
          className="input-ex"
          value={ex}
          onChange={(e) => setEx(e.target.value)}
          title="durée de vie optionnelle, en secondes"
        />
        <button type="submit" disabled={disabled || key.trim() === ""}>
          SET
        </button>
      </form>

      <div className="toolbar-group">
        <button disabled={disabled} onClick={() => void seed(CONFIG.seedSmall)}>
          Seed {CONFIG.seedSmall.toLocaleString("fr-FR")}
        </button>
        <button disabled={disabled} onClick={() => void seed(CONFIG.seedLarge)}>
          Seed {CONFIG.seedLarge.toLocaleString("fr-FR")}
        </button>
        {seedProgress !== null && (
          <progress value={seedProgress} max={1} title={`${Math.round(seedProgress * 100)} %`} />
        )}
      </div>

      <div className="toolbar-group">
        <button className="highlight" disabled={disabled} onClick={() => void touchRandomEntry()}>
          🎯 Modifier 1 ligne au hasard
        </button>
      </div>

      <QueryTester disabled={disabled} />

      <div className="toolbar-group">
        <button disabled={disabled} onClick={() => void flushNow()} title="buffer → aof.log, sans attendre le tick">
          Flush
        </button>
        <button disabled={disabled} onClick={() => void snapshotNow()} title="photo complète + compaction de l'AOF">
          Snapshot
        </button>
        {busy !== null && <span className="busy">⏳ {busy}…</span>}
      </div>
    </div>
  );
}

function QueryTester({ disabled }: { disabled: boolean }) {
  const result = useSelector(() => store$.queryResult.get());
  const [op, setOp] = useState<FilterOp>(">");
  const [pivot, setPivot] = useState("50000");

  return (
    <div className="toolbar-group query-tester">
      <span className="query-label">GET</span>
      <select name="op" value={op} onChange={(e) => setOp(e.target.value as FilterOp)}>
        {["=", ">", ">=", "<", "<=", "contains"].map((o) => (
          <option key={o} value={o}>
            {o}
          </option>
        ))}
      </select>
      <input name="pivot" value={pivot} onChange={(e) => setPivot(e.target.value)} placeholder="pivot" />
      <button disabled={disabled || pivot === ""} onClick={() => void runQuery(op, pivot)}>
        exec
      </button>
      {result && (
        <span className="query-result" title={result.keys.slice(0, 30).join(", ")}>
          {result.label} → {result.keys.length.toLocaleString("fr-FR")} clé(s) en {result.tookMs} ms
        </span>
      )}
    </div>
  );
}
