import { useSelector } from "@legendapp/state/react";
import { useEffect } from "react";
import { initApp } from "./actions";
import { store$ } from "./store";
import { Toolbar } from "./Toolbar";
import { VirtualList } from "./VirtualList";

export function App() {
  const ready = useSelector(() => store$.ready.get());
  const toast = useSelector(() => store$.toast.get());

  useEffect(() => {
    void initApp();
  }, []);

  return (
    <div className="app">
      <header>
        <h1>WasmRedis</h1>
        <p className="subtitle">
          moteur Go → WASM · persistance OPFS · SDK TypeScript · virtual scroll fait main ·
          render granulaire Legend State (la colonne de droite compte les renders de chaque ligne)
        </p>
      </header>

      {ready ? (
        <>
          <Toolbar />
          <VirtualList />
        </>
      ) : (
        <p className="loading">⏳ démarrage du moteur (worker + WASM + restore OPFS)…</p>
      )}

      {toast && <div className={`toast toast-${toast.kind}`}>{toast.text}</div>}
    </div>
  );
}
