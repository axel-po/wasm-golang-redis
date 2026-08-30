interface SyncAccessHandle {
  read(buffer: Uint8Array, options?: { at: number }): number;
  write(buffer: Uint8Array, options?: { at: number }): number;
  truncate(newSize: number): void;
  getSize(): number;
  flush(): void;
  close(): void;
}

declare global {
  interface FileSystemFileHandle {
    createSyncAccessHandle(): Promise<SyncAccessHandle>;
  }
}

export interface StorageBridge {
  appendAOF: (text: string) => void;
  readAOF: () => string;
  clearAOF: () => void;
  writeSnapshot: (json: string) => void;
  readSnapshot: () => string;
}

const encoder = new TextEncoder();
const decoder = new TextDecoder();

const readAll = (handle: SyncAccessHandle): string => {
  const buffer = new Uint8Array(handle.getSize());
  handle.read(buffer, { at: 0 });
  return decoder.decode(buffer);
};

const overwrite = (handle: SyncAccessHandle, text: string): void => {
  handle.truncate(0);
  handle.write(encoder.encode(text), { at: 0 });
  handle.flush();
};

const createMemoryStorage = (): StorageBridge => {
  let aof = "";
  let snapshot = "";
  return {
    appendAOF: (text) => {
      aof += text;
    },
    readAOF: () => aof,
    clearAOF: () => {
      aof = "";
    },
    writeSnapshot: (json) => {
      snapshot = json;
    },
    readSnapshot: () => snapshot,
  };
};

export const createOpfsStorage = async (): Promise<StorageBridge> => {
  try {
    return await openOpfsStorage();
  } catch (err) {

    console.warn("OPFS indisponible — persistance désactivée sur cet onglet:", err);
    return createMemoryStorage();
  }
};

const openOpfsStorage = async (): Promise<StorageBridge> => {
  const root = await navigator.storage.getDirectory();
  const dir = await root.getDirectoryHandle("wasmredis", { create: true });

  const open = async (name: string): Promise<SyncAccessHandle> => {
    const file = await dir.getFileHandle(name, { create: true });
    let lastErr: unknown;
    for (let attempt = 0; attempt < 10; attempt++) {
      try {
        return await file.createSyncAccessHandle();
      } catch (err) {
        lastErr = err;
        await new Promise((resolve) => setTimeout(resolve, 150));
      }
    }
    throw lastErr;
  };

  const aof = await open("aof.log");
  const snapshot = await open("snapshot.json");

  return {

    appendAOF: (text) => {
      aof.write(encoder.encode(text), { at: aof.getSize() });
      aof.flush();
    },
    readAOF: () => readAll(aof),

    clearAOF: () => {
      aof.truncate(0);
      aof.flush();
    },

    writeSnapshot: (json) => overwrite(snapshot, json),
    readSnapshot: () => readAll(snapshot),
  };
};
