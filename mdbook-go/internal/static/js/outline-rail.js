/* ============================================================
   右侧大纲栏 —— use-active-headings.ts + section-rail.tsx 的原生版
   ============================================================
   Go-only（Rust 端无对应物）。单文件移植自
   mdbook-go/test-html-css/index.html，数据源是现成 DOM：
   main.markdown-body 里的 h1..h6[id]（add_header_links 已生成）。

   交互：hover 右缘 tick 条 → 弹出大纲卡片；点击跳转；滚动高亮当前
   标题；Esc 关闭；右键复制 [标题](#锚点) 链接。
   ============================================================ */

(function () {
    const ACTIVE_OFFSET_PX = 144;  // --heading-scroll-offset（writer.css），与 :target 一致
    const TICK_HEIGHT = 1;
    const TICK_GAP = 6;
    const POPOVER_TRANSITION_MS = 180;

    const prose = document.querySelector("main.markdown-body");
    if (!prose) return;

    const headings = Array.from(prose.querySelectorAll("h1, h2, h3, h4, h5, h6")).map((el) => ({
        el,
        text: el.textContent.trim(),
        slug: el.id,
    }));
    if (headings.length === 0) return;

    const railTicks = document.getElementById("rail-ticks");
    const railTrigger = document.getElementById("rail-trigger");
    const railPopover = document.getElementById("rail-popover");
    const railPopoverList = document.getElementById("rail-popover-list");
    if (!railTicks || !railTrigger || !railPopover || !railPopoverList) return;

    let isOpen = false;
    let activeIndex = 0;

    /* 构建 tick 条与大纲行 */
    function buildRail() {
        const tickStackHeight =
            headings.length * TICK_HEIGHT + Math.max(0, headings.length - 1) * TICK_GAP;
        railTrigger.style.height = tickStackHeight + "px";

        headings.forEach((h, i) => {
            const tick = document.createElement("button");
            tick.type = "button";
            tick.className = "rail-tick" + (i === 0 ? " is-active" : "");
            tick.style.width = "20px";
            tick.style.height = TICK_HEIGHT + "px";
            tick.title = h.text;
            tick.setAttribute("aria-label", h.text);
            tick.addEventListener("click", () => scrollToHeading(i));
            tick.addEventListener("contextmenu", (e) => openMenu(e, i));
            railTicks.appendChild(tick);

            const row = document.createElement("button");
            row.type = "button";
            row.className = "rail-popover-row" + (i === 0 ? " is-active" : "");
            row.textContent = h.text;
            row.title = h.text;
            row.addEventListener("click", () => scrollToHeading(i));
            row.addEventListener("contextmenu", (e) => openMenu(e, i));
            railPopoverList.appendChild(row);
        });
    }

    /* 计算当前激活标题：最后一个顶边越过阈值的标题（Writer 同款逻辑） */
    function computeActive() {
        const threshold = ACTIVE_OFFSET_PX;
        let idx = null;
        for (let i = 0; i < headings.length; i++) {
            if (headings[i].el.getBoundingClientRect().top > threshold) break;
            idx = i;
        }
        return idx === null ? 0 : idx;
    }

    function applyActive(idx) {
        if (idx === activeIndex) return;
        activeIndex = idx;
        railTicks.querySelectorAll(".rail-tick").forEach((t, i) =>
            t.classList.toggle("is-active", i === idx));
        railPopoverList.querySelectorAll(".rail-popover-row").forEach((r, i) =>
            r.classList.toggle("is-active", i === idx));
    }

    /* 滚动监听 —— rAF 节流（use-active-headings.ts） */
    let frame = 0;
    function scheduleActive() {
        if (frame) return;
        frame = requestAnimationFrame(() => {
            frame = 0;
            applyActive(computeActive());
        });
    }
    window.addEventListener("scroll", scheduleActive, { passive: true });
    window.addEventListener("resize", scheduleActive);
    scheduleActive();

    /* 跳转 —— 与 Writer 一致：落在 ACTIVE_OFFSET_PX 处（scrollToHeading） */
    function scrollToHeading(i) {
        const top = headings[i].el.getBoundingClientRect().top + window.scrollY - ACTIVE_OFFSET_PX;
        window.scrollTo({ top: Math.max(0, top), behavior: "auto" });
        applyActive(i);
    }

    /* 开合 —— hover 进出 + relatedTarget 桥接（rail ↔ popover 有重叠区） */
    function openRail() {
        isOpen = true;
        railTicks.dataset.open = "true";
        railPopover.hidden = false;
        requestAnimationFrame(() => { railPopover.dataset.state = "open"; });
    }
    function closeRail() {
        if (!isOpen) return;
        isOpen = false;
        railTicks.dataset.open = "false";
        railPopover.dataset.state = "closed";
        setTimeout(() => {
            if (!isOpen) railPopover.hidden = true;
        }, POPOVER_TRANSITION_MS);
    }
    const bridges = (e, ref) =>
        e.relatedTarget instanceof Node && ref.contains(e.relatedTarget);

    railTrigger.addEventListener("mouseenter", openRail);
    railTrigger.addEventListener("mouseleave", (e) => {
        if (!bridges(e, railPopover)) closeRail();
    });
    railPopover.addEventListener("mouseenter", () => { if (isOpen) return; });
    railPopover.addEventListener("mouseleave", (e) => {
        if (!bridges(e, railTrigger)) closeRail();
    });

    /* Esc 关闭（useEscKey） */
    document.addEventListener("keydown", (e) => {
        if (e.key === "Escape") { closeRail(); closeMenu(); }
    });

    /* 右键菜单 —— 复制标题链接（[标题](#slug)） */
    const menu = document.createElement("div");
    menu.className = "rail-menu";
    menu.hidden = true;
    document.body.appendChild(menu);

    function openMenu(e, i) {
        e.preventDefault();
        e.stopPropagation();
        const h = headings[i];
        const btn = document.createElement("button");
        btn.textContent = "复制标题链接";
        btn.addEventListener("click", () => {
            const link = `[${h.text}](#${h.slug})`;
            navigator.clipboard
                .writeText(link)
                .catch(() => {
                    const ta = document.createElement("textarea");
                    ta.value = link;
                    document.body.appendChild(ta);
                    ta.select();
                    document.execCommand("copy");
                    ta.remove();
                })
                .finally(() => {
                    btn.textContent = "已复制 ✓";
                    setTimeout(closeMenu, 800);
                });
        });
        menu.replaceChildren(btn);
        const x = Math.min(e.clientX, window.innerWidth - 180);
        const y = Math.min(e.clientY, window.innerHeight - 44);
        menu.style.left = x + "px";
        menu.style.top = y + "px";
        menu.hidden = false;
    }
    function closeMenu() { menu.hidden = true; }
    document.addEventListener("click", (e) => {
        if (!menu.contains(e.target)) closeMenu();
    });
    window.addEventListener("blur", closeMenu);

    buildRail();
})();
