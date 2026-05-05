#!/usr/bin/env python3
from __future__ import annotations

import json
import random
import threading
import tkinter as tk
import urllib.error
from pathlib import Path
from tkinter import filedialog, messagebox, ttk

SCRIPT_DIR = Path(__file__).resolve().parent
import sys
if str(SCRIPT_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPT_DIR))

from proxy_node_core import DEFAULT_URL, SUPPORTED_PROTOCOLS, fetch_and_normalize

BG = "#070B14"
PANEL = "#0D1322"
PANEL_2 = "#111A2D"
TEXT = "#E8F7FF"
MUTED = "#8EA7C2"
ACCENT = "#00E5FF"
ACCENT_2 = "#8A5CFF"
ACCENT_3 = "#22FFA5"
WARN = "#FF5AA5"
GRID = "#18233A"


class NeonButton(tk.Canvas):
    def __init__(self, master, text: str, command, width: int = 168, height: int = 44, colors: tuple[str, str] = (ACCENT, ACCENT_2)):
        super().__init__(master, width=width, height=height, bg=master.cget("bg"), highlightthickness=0, bd=0)
        self.command = command
        self.colors = colors
        self.width = width
        self.height = height
        self.text = text
        self.enabled = True
        self.bind("<Button-1>", self._clicked)
        self.bind("<Enter>", lambda _e: self.draw(hover=True))
        self.bind("<Leave>", lambda _e: self.draw(hover=False))
        self.draw(False)

    def set_enabled(self, enabled: bool) -> None:
        self.enabled = enabled
        self.draw(False)

    def _clicked(self, _event) -> None:
        if self.enabled:
            self.command()

    def draw(self, hover: bool) -> None:
        self.delete("all")
        fill = self.colors[0] if self.enabled else "#334155"
        outline = self.colors[1] if self.enabled else "#475569"
        shadow = outline if hover and self.enabled else fill
        self.create_rectangle(9, 9, self.width - 1, self.height - 1, outline="", fill=shadow)
        self.create_rectangle(0, 0, self.width - 10, self.height - 10, outline=outline, width=2, fill=fill)
        self.create_text(
            (self.width - 10) / 2,
            (self.height - 10) / 2,
            text=self.text,
            fill="#03131B" if self.enabled else "#94A3B8",
            font=("Segoe UI", 11, "bold"),
        )


class Starfield(tk.Canvas):
    def __init__(self, master):
        super().__init__(master, bg=BG, highlightthickness=0, bd=0)
        self.particles: list[dict] = []
        self.bind("<Configure>", self._on_resize)
        self.after(60, self.animate)

    def _on_resize(self, _event) -> None:
        if not self.particles:
            self._seed_particles()
        self.redraw()

    def _seed_particles(self) -> None:
        w = max(self.winfo_width(), 1200)
        h = max(self.winfo_height(), 780)
        self.particles = []
        for _ in range(55):
            self.particles.append(
                {
                    "x": random.uniform(0, w),
                    "y": random.uniform(0, h),
                    "r": random.uniform(1.0, 2.8),
                    "dx": random.uniform(-0.25, 0.25),
                    "dy": random.uniform(0.12, 0.52),
                    "color": random.choice((ACCENT, ACCENT_2, ACCENT_3, "#FFFFFF")),
                }
            )

    def redraw(self) -> None:
        self.delete("all")
        w = self.winfo_width() or 1200
        h = self.winfo_height() or 780
        self.create_rectangle(0, 0, w, h, fill=BG, outline="")
        for i in range(0, w, 44):
            self.create_line(i, 0, i, h, fill=GRID)
        for j in range(0, h, 44):
            self.create_line(0, j, w, j, fill=GRID)
        self.create_arc(-280, -220, 420, 430, start=10, extent=190, style="arc", outline="#12395B", width=2)
        self.create_arc(w - 520, h - 400, w + 120, h + 140, start=200, extent=160, style="arc", outline="#251F5C", width=2)
        for particle in self.particles:
            x, y, r = particle["x"], particle["y"], particle["r"]
            self.create_oval(x - r * 2.6, y - r * 2.6, x + r * 2.6, y + r * 2.6, fill=particle["color"], outline="")
            self.create_oval(x - r, y - r, x + r, y + r, fill="#FFFFFF", outline="")

    def animate(self) -> None:
        if not self.particles:
            self._seed_particles()
        w = self.winfo_width() or 1200
        h = self.winfo_height() or 780
        for particle in self.particles:
            particle["x"] = (particle["x"] + particle["dx"]) % w
            particle["y"] = (particle["y"] + particle["dy"]) % h
        self.redraw()
        self.after(60, self.animate)


class ProxyNodeStudio:
    def __init__(self, root: tk.Tk):
        self.root = root
        self.root.title("Proxy Node Studio")
        self.root.geometry("1380x860")
        self.root.minsize(1180, 760)
        self.root.configure(bg=BG)
        self.root.option_add("*tearOff", False)

        self.nodes: list[dict] = []
        self.last_result: dict | None = None
        self.protocol_vars = {protocol: tk.BooleanVar(value=True) for protocol in SUPPORTED_PROTOCOLS}
        self.url_var = tk.StringVar(value=DEFAULT_URL)
        self.status_var = tk.StringVar(value="Ready. 点击 Quantum Fetch 开始抓取节点。")
        self.total_var = tk.StringVar(value="0")
        self.protocol_count_vars = {protocol: tk.StringVar(value="0") for protocol in SUPPORTED_PROTOCOLS}
        self.output_path_var = tk.StringVar(value="未导出")

        self._build_styles()
        self._build_layout()

    def _build_styles(self) -> None:
        style = ttk.Style()
        style.theme_use("clam")
        style.configure("Treeview", background=PANEL, foreground=TEXT, fieldbackground=PANEL, borderwidth=0, rowheight=30)
        style.map("Treeview", background=[("selected", ACCENT_2)], foreground=[("selected", "#FFFFFF")])
        style.configure("Treeview.Heading", background="#0B1E31", foreground="#DDFBFF", relief="flat", font=("Segoe UI", 10, "bold"))
        style.configure("Studio.Horizontal.TProgressbar", troughcolor="#09111F", background=ACCENT, bordercolor="#09111F", lightcolor=ACCENT, darkcolor=ACCENT)

    def _build_layout(self) -> None:
        self.bg_canvas = Starfield(self.root)
        self.bg_canvas.place(relx=0, rely=0, relwidth=1, relheight=1)

        self.overlay = tk.Frame(self.root, bg=BG)
        self.overlay.pack(fill="both", expand=True, padx=18, pady=18)
        self.overlay.grid_columnconfigure(0, weight=1)
        self.overlay.grid_rowconfigure(2, weight=1)

        self._build_header()
        self._build_control_panel()
        self._build_results_panel()

    def _card(self, master, pad=14):
        frame = tk.Frame(master, bg=PANEL, highlightbackground="#12314D", highlightthickness=1)
        frame.pack_propagate(False)
        inner = tk.Frame(frame, bg=PANEL)
        inner.pack(fill="both", expand=True, padx=pad, pady=pad)
        return frame, inner

    def _build_header(self) -> None:
        frame, inner = self._card(self.overlay, pad=16)
        frame.grid(row=0, column=0, sticky="ew", pady=(0, 14))
        inner.grid_columnconfigure(1, weight=1)

        logo = tk.Canvas(inner, width=90, height=90, bg=PANEL, highlightthickness=0)
        logo.grid(row=0, column=0, rowspan=2, padx=(0, 18))
        for radius, color in ((40, ACCENT_2), (32, ACCENT), (22, ACCENT_3)):
            logo.create_oval(45 - radius, 45 - radius, 45 + radius, 45 + radius, outline=color, width=2)
        logo.create_line(15, 45, 75, 45, fill="#FFFFFF", width=2)
        logo.create_line(45, 15, 45, 75, fill="#FFFFFF", width=2)
        logo.create_text(45, 45, text="PX", fill="#FFFFFF", font=("Segoe UI", 16, "bold"))

        tk.Label(inner, text="PROXY NODE STUDIO", bg=PANEL, fg=ACCENT, font=("Segoe UI", 28, "bold")).grid(row=0, column=1, sticky="w")
        tk.Label(
            inner,
            text="Cyber UI / 节点抓取 / JSON 导出 / 支持 SS · VMESS · VLESS · TROJAN",
            bg=PANEL,
            fg=MUTED,
            font=("Segoe UI", 11),
        ).grid(row=1, column=1, sticky="w")

    def _build_control_panel(self) -> None:
        grid = tk.Frame(self.overlay, bg=BG)
        grid.grid(row=1, column=0, sticky="ew", pady=(0, 14))
        for i in range(5):
            grid.grid_columnconfigure(i, weight=1)

        control_frame, control = self._card(grid)
        control_frame.grid(row=0, column=0, columnspan=3, sticky="nsew", padx=(0, 10))
        tk.Label(control, text="SOURCE MATRIX", bg=PANEL, fg=TEXT, font=("Segoe UI", 15, "bold")).pack(anchor="w")
        tk.Label(control, text="目标 JSON 地址", bg=PANEL, fg=MUTED, font=("Segoe UI", 10)).pack(anchor="w", pady=(10, 4))
        self.url_entry = tk.Entry(control, textvariable=self.url_var, bg="#08111F", fg=TEXT, insertbackground=ACCENT, relief="flat", font=("Consolas", 11))
        self.url_entry.pack(fill="x", ipady=8)

        protocols_row = tk.Frame(control, bg=PANEL)
        protocols_row.pack(fill="x", pady=(12, 10))
        for protocol in SUPPORTED_PROTOCOLS:
            chip = tk.Checkbutton(
                protocols_row,
                text=protocol.upper(),
                variable=self.protocol_vars[protocol],
                bg=PANEL,
                fg=TEXT,
                selectcolor="#0B1E31",
                activebackground=PANEL,
                activeforeground=ACCENT,
                font=("Segoe UI", 10, "bold"),
                padx=12,
                pady=6,
            )
            chip.pack(side="left", padx=(0, 8))

        buttons = tk.Frame(control, bg=PANEL)
        buttons.pack(fill="x", pady=(4, 0))
        self.fetch_button = NeonButton(buttons, "Quantum Fetch", self.fetch_nodes, colors=(ACCENT, ACCENT_2))
        self.fetch_button.pack(side="left", padx=(0, 10))
        self.export_button = NeonButton(buttons, "Export JSON", self.export_json, colors=(ACCENT_3, ACCENT))
        self.export_button.pack(side="left", padx=(0, 10))
        self.copy_button = NeonButton(buttons, "Copy URI", self.copy_selected_uri, colors=(WARN, ACCENT_2))
        self.copy_button.pack(side="left")

        self.progress = ttk.Progressbar(control, style="Studio.Horizontal.TProgressbar", mode="indeterminate")
        self.progress.pack(fill="x", pady=(14, 10))
        tk.Label(control, textvariable=self.status_var, bg=PANEL, fg=MUTED, font=("Segoe UI", 10)).pack(anchor="w")

        total_frame, total_inner = self._card(grid)
        total_frame.grid(row=0, column=3, sticky="nsew", padx=(0, 10))
        tk.Label(total_inner, text="TOTAL NODES", bg=PANEL, fg=MUTED, font=("Segoe UI", 11, "bold")).pack(anchor="w")
        tk.Label(total_inner, textvariable=self.total_var, bg=PANEL, fg=ACCENT, font=("Segoe UI", 34, "bold")).pack(anchor="w", pady=(8, 8))
        tk.Label(total_inner, textvariable=self.output_path_var, bg=PANEL, fg="#6EE7B7", font=("Segoe UI", 9), wraplength=220, justify="left").pack(anchor="w")

        stats_frame, stats_inner = self._card(grid)
        stats_frame.grid(row=0, column=4, sticky="nsew")
        tk.Label(stats_inner, text="PROTOCOL COUNTS", bg=PANEL, fg=MUTED, font=("Segoe UI", 11, "bold")).pack(anchor="w")
        for protocol, color in zip(SUPPORTED_PROTOCOLS, (ACCENT, ACCENT_2, ACCENT_3, WARN)):
            row = tk.Frame(stats_inner, bg=PANEL)
            row.pack(fill="x", pady=4)
            tk.Label(row, text=protocol.upper(), bg=PANEL, fg=color, font=("Segoe UI", 11, "bold")).pack(side="left")
            tk.Label(row, textvariable=self.protocol_count_vars[protocol], bg=PANEL, fg=TEXT, font=("Consolas", 12, "bold")).pack(side="right")

    def _build_results_panel(self) -> None:
        frame, inner = self._card(self.overlay)
        frame.grid(row=2, column=0, sticky="nsew")
        inner.grid_rowconfigure(1, weight=1)
        inner.grid_columnconfigure(0, weight=1)

        tk.Label(inner, text="NODE GRID", bg=PANEL, fg=TEXT, font=("Segoe UI", 15, "bold")).grid(row=0, column=0, sticky="w")

        columns = ("protocol", "name", "host", "port", "network", "tls")
        self.tree = ttk.Treeview(inner, columns=columns, show="headings", selectmode="browse")
        headings = {
            "protocol": "Protocol",
            "name": "Node Name",
            "host": "Host",
            "port": "Port",
            "network": "Network",
            "tls": "TLS",
        }
        widths = {
            "protocol": 110,
            "name": 250,
            "host": 330,
            "port": 90,
            "network": 120,
            "tls": 120,
        }
        for key in columns:
            self.tree.heading(key, text=headings[key])
            self.tree.column(key, width=widths[key], anchor="center")
        self.tree.grid(row=1, column=0, sticky="nsew", pady=(10, 0))
        self.tree.bind("<<TreeviewSelect>>", lambda _e: self._refresh_preview())

        scrollbar = ttk.Scrollbar(inner, orient="vertical", command=self.tree.yview)
        scrollbar.grid(row=1, column=1, sticky="ns", pady=(10, 0))
        self.tree.configure(yscrollcommand=scrollbar.set)

        lower = tk.Frame(inner, bg=PANEL)
        lower.grid(row=2, column=0, columnspan=2, sticky="ew", pady=(14, 0))
        lower.grid_columnconfigure(0, weight=1)
        lower.grid_columnconfigure(1, weight=1)

        preview_frame, preview_inner = self._card(lower, pad=12)
        preview_frame.grid(row=0, column=0, sticky="nsew", padx=(0, 8))
        tk.Label(preview_inner, text="RAW URI PREVIEW", bg=PANEL, fg=MUTED, font=("Segoe UI", 11, "bold")).pack(anchor="w")
        self.uri_preview = tk.Text(preview_inner, height=5, bg="#08111F", fg=TEXT, insertbackground=ACCENT, relief="flat", font=("Consolas", 10), wrap="word")
        self.uri_preview.pack(fill="both", expand=True, pady=(8, 0))

        log_frame, log_inner = self._card(lower, pad=12)
        log_frame.grid(row=0, column=1, sticky="nsew", padx=(8, 0))
        tk.Label(log_inner, text="SYSTEM LOG", bg=PANEL, fg=MUTED, font=("Segoe UI", 11, "bold")).pack(anchor="w")
        self.log_text = tk.Text(log_inner, height=5, bg="#08111F", fg="#B6F6FF", insertbackground=ACCENT, relief="flat", font=("Consolas", 10), wrap="word")
        self.log_text.pack(fill="both", expand=True, pady=(8, 0))
        self._log("Studio initialized. Waiting for fetch command.")

    def _selected_protocols(self) -> tuple[str, ...]:
        selected = tuple(protocol for protocol, var in self.protocol_vars.items() if var.get())
        return selected or SUPPORTED_PROTOCOLS

    def _log(self, message: str) -> None:
        self.log_text.insert("end", message + "\n")
        self.log_text.see("end")

    def _set_busy(self, busy: bool) -> None:
        self.fetch_button.set_enabled(not busy)
        self.export_button.set_enabled(not busy)
        self.copy_button.set_enabled(not busy)
        if busy:
            self.progress.start(12)
        else:
            self.progress.stop()

    def fetch_nodes(self) -> None:
        url = self.url_var.get().strip()
        if not url:
            messagebox.showwarning("Missing URL", "请输入 list.json 地址。")
            return
        protocols = self._selected_protocols()
        self.status_var.set("Fetching remote payload...")
        self._log(f"[FETCH] url={url} protocols={','.join(protocols)}")
        self._set_busy(True)

        def worker() -> None:
            try:
                result = fetch_and_normalize(url=url, timeout=20.0, protocols=protocols)
                self.root.after(0, lambda: self._apply_result(result))
            except urllib.error.URLError as exc:
                self.root.after(0, lambda: self._handle_error(f"网络错误: {exc}"))
            except json.JSONDecodeError as exc:
                self.root.after(0, lambda: self._handle_error(f"JSON 解析失败: {exc}"))
            except Exception as exc:
                self.root.after(0, lambda: self._handle_error(f"未知错误: {exc}"))

        threading.Thread(target=worker, daemon=True).start()

    def _apply_result(self, result: dict) -> None:
        self.nodes = result["nodes"]
        self.last_result = result
        self.output_path_var.set("未导出")
        self.total_var.set(str(result["total_nodes"]))
        for protocol in SUPPORTED_PROTOCOLS:
            self.protocol_count_vars[protocol].set(str(result["protocol_counts"].get(protocol, 0)))
        self.tree.delete(*self.tree.get_children())
        for index, node in enumerate(self.nodes):
            self.tree.insert("", "end", iid=str(index), values=(node["protocol"].upper(), node["name"], node["host"], node["port"], node["network"], node["tls"]))
        self.status_var.set(f"Fetch complete. {result['total_nodes']} nodes loaded.")
        self._log(f"[OK] loaded {result['total_nodes']} nodes")
        self._refresh_preview()
        self._set_busy(False)
        if self.nodes:
            self.tree.selection_set("0")
            self.tree.focus("0")
            self._refresh_preview()

    def _handle_error(self, message: str) -> None:
        self.status_var.set(message)
        self._log(f"[ERROR] {message}")
        self._set_busy(False)
        messagebox.showerror("Fetch Failed", message)

    def _refresh_preview(self) -> None:
        self.uri_preview.delete("1.0", "end")
        selection = self.tree.selection()
        if not selection:
            return
        node = self.nodes[int(selection[0])]
        self.uri_preview.insert("1.0", node.get("raw_uri", ""))

    def copy_selected_uri(self) -> None:
        selection = self.tree.selection()
        if not selection:
            messagebox.showinfo("No Selection", "请先在列表中选一个节点。")
            return
        node = self.nodes[int(selection[0])]
        self.root.clipboard_clear()
        self.root.clipboard_append(node.get("raw_uri", ""))
        self.status_var.set("Selected URI copied to clipboard.")
        self._log(f"[COPY] {node.get('host')}:{node.get('port')}")

    def export_json(self) -> None:
        if not self.last_result:
            messagebox.showinfo("No Data", "先抓取节点，再导出 JSON。")
            return
        target = filedialog.asksaveasfilename(
            title="Export node JSON",
            defaultextension=".json",
            filetypes=[("JSON Files", "*.json"), ("All Files", "*.*")],
            initialfile="proxy_nodes_studio.json",
        )
        if not target:
            return
        path = Path(target)
        path.write_text(json.dumps(self.last_result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        self.output_path_var.set(str(path))
        self.status_var.set(f"Exported to {path.name}")
        self._log(f"[EXPORT] {path}")


def main() -> None:
    root = tk.Tk()
    ProxyNodeStudio(root)
    root.mainloop()


if __name__ == "__main__":
    main()
