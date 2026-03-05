import os, json, pathlib
import numpy as np
import soundfile as sf
import openl3

IN_DIR  = os.environ.get("DAW_IN_DIR", "/work/in")
OUT_DIR = os.environ.get("DAW_OUT_DIR", "/work/out")

content_type   = os.environ.get("OPENL3_CONTENT_TYPE", "music")
input_repr     = os.environ.get("OPENL3_INPUT_REPR", "mel256")
embedding_size = int(os.environ.get("OPENL3_EMBEDDING_SIZE", "6144"))
hop_size       = float(os.environ.get("OPENL3_HOP_SIZE", "0.1"))
batch_size     = int(os.environ.get("OPENL3_BATCH_SIZE", "8"))

pathlib.Path(OUT_DIR).mkdir(parents=True, exist_ok=True)

def is_audio(p: pathlib.Path) -> bool:
    return p.suffix.lower() in [".wav",".flac",".ogg",".mp3",".m4a",".aac",".aiff",".aif"]

files = [p for p in pathlib.Path(IN_DIR).iterdir() if p.is_file() and is_audio(p)]
files.sort()

for p in files:
    stem = p.stem
    out_emb = pathlib.Path(OUT_DIR) / f"{stem}.openl3.npy"
    out_ts  = pathlib.Path(OUT_DIR) / f"{stem}.openl3.timestamps.npy"
    out_meta = pathlib.Path(OUT_DIR) / f"{stem}.openl3.json"

    audio, sr = sf.read(str(p), always_2d=False)
    if isinstance(audio, np.ndarray) and audio.ndim == 2:
        audio = np.mean(audio, axis=1)

    emb, ts = openl3.get_audio_embedding(
        audio,
        sr,
        content_type=content_type,
        input_repr=input_repr,
        embedding_size=embedding_size,
        hop_size=hop_size,
        batch_size=batch_size,
    )

    np.save(out_emb, emb)
    np.save(out_ts, ts)

    meta = {
        "file": p.name,
        "sr": int(sr),
        "content_type": content_type,
        "input_repr": input_repr,
        "embedding_size": embedding_size,
        "hop_size": hop_size,
        "batch_size": batch_size,
        "frames": int(emb.shape[0]),
        "dims": int(emb.shape[1]),
    }
    out_meta.write_text(json.dumps(meta, indent=2))

    print(f"ok: {p.name} -> {out_emb.name} ({emb.shape[0]}x{emb.shape[1]})")

print("done")
