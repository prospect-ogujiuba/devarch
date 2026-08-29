const state = {
  inventory: { projects: [], containers: [], services: [], runtimeError: null },
  query: "",
  loaded: false,
};

const elements = {
  view: document.querySelector("#view"),
  search: document.querySelector("#search"),
  searchPanel: document.querySelector("#search-panel"),
  refreshButton: document.querySelector("#refresh-button"),
  refreshStatus: document.querySelector("#refresh-status"),
  runtimeAlert: document.querySelector("#runtime-alert"),
  pageEyebrow: document.querySelector("#page-eyebrow"),
  pageTitle: document.querySelector("#page-title"),
  pageDescription: document.querySelector("#page-description"),
  pageBack: document.querySelector("#page-back"),
  menuButton: document.querySelector("#menu-button"),
  mobileMenu: document.querySelector("#mobile-menu"),
  menuOpenIcon: document.querySelector("#menu-open-icon"),
  menuCloseIcon: document.querySelector("#menu-close-icon"),
  toast: document.querySelector("#toast"),
};

function node(tag, className = "", text = "") {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (text) element.textContent = text;
  return element;
}

function routeLink(label, href, className = "action") {
  const link = node("a", className, label);
  link.href = href;
  link.dataset.route = "";
  return link;
}

function action(label, options = {}) {
  const element = node(options.href ? "a" : "button", options.primary ? "primary-action" : "action", label);
  if (options.href) {
    element.href = options.href;
    if (options.external) {
      element.target = "_blank";
      element.rel = "noreferrer";
    }
  } else {
    element.type = "button";
  }
  if (options.title) element.title = options.title;
  if (options.onClick) element.addEventListener("click", options.onClick);
  return element;
}

function appPath(project) {
  return `/apps/${encodeURIComponent(project.name)}`;
}

function servicePath(service) {
  return `/services/${encodeURIComponent(service.category)}/${encodeURIComponent(service.name)}`;
}

function searchable(item) {
  return Object.values(item)
    .filter((value) => typeof value === "string")
    .join(" ")
    .toLocaleLowerCase();
}

function filtered(items) {
  if (!state.query) return items;
  return items.filter((item) => searchable(item).includes(state.query));
}

function emptyState(message) {
  return node("div", "col-span-full rounded-xl border border-dashed border-white/10 px-5 py-10 text-center text-sm text-slate-500", message);
}

function sectionHeader(eyebrow, title, count, href) {
  const header = node("div", "mb-4 flex items-end justify-between gap-4");
  const labels = node("div");
  labels.append(
    node("p", "text-xs font-bold uppercase tracking-[0.2em] text-teal-300", eyebrow),
    node("h2", "mt-1 text-xl font-bold text-white", title),
  );
  const right = node("div", "flex items-center gap-3");
  right.append(node("span", "badge", String(count)));
  if (href) right.append(routeLink("View all", href));
  header.append(labels, right);
  return header;
}

let toastTimer;
function showToast(message) {
  window.clearTimeout(toastTimer);
  elements.toast.textContent = message;
  elements.toast.classList.remove("hidden");
  toastTimer = window.setTimeout(() => elements.toast.classList.add("hidden"), 1800);
}

async function copyText(value, label) {
  try {
    await navigator.clipboard.writeText(value);
    showToast(`${label} copied`);
  } catch (_error) {
    showToast("Clipboard access was blocked");
  }
}

function projectCard(project) {
  const card = node("article", "surface group flex min-h-52 flex-col p-5 transition hover:-translate-y-0.5 hover:border-teal-300/25");
  const top = node("div", "flex items-start justify-between gap-4");
  const titleWrap = node("div", "min-w-0");
  const title = routeLink(project.name, appPath(project), "block truncate text-lg font-bold text-white hover:text-teal-200 focus:outline-none focus:ring-2 focus:ring-teal-400/60");
  titleWrap.append(title, node("p", "mt-1 truncate text-xs text-slate-500", project.relativePath));
  top.append(titleWrap, node("span", "badge shrink-0", project.kind));

  const url = node("p", "mt-5 truncate font-mono text-xs text-teal-300", project.url);
  const controls = node("div", "mt-auto flex flex-wrap gap-2 pt-5");
  controls.append(
    routeLink("Details", appPath(project)),
    action("Open site", { href: project.url, external: true }),
    action("Copy path", { onClick: () => copyText(project.path, "Path") }),
  );
  card.append(top, url, controls);
  return card;
}

function containerRow(container) {
  const row = node("article", "flex flex-col gap-4 px-4 py-4 sm:flex-row sm:items-center sm:px-5");
  const status = node("span", "size-2.5 shrink-0 rounded-full bg-emerald-400 shadow-[0_0_12px_rgba(52,211,153,.65)]");
  const info = node("div", "min-w-0 flex-1");
  const heading = node("div", "flex min-w-0 items-center gap-3");
  heading.append(status, node("h3", "truncate font-bold text-white", container.name));
  info.append(heading, node("p", "mt-1 truncate text-xs text-slate-500", `${container.image} · ${container.status || container.state}`));

  const portText = container.ports.length
    ? container.ports.map((port) => `${port.hostPort}:${port.containerPort ?? "?"}`).join(", ")
    : "No published ports";
  const ports = node("p", "break-all font-mono text-xs text-slate-400 sm:w-48 sm:text-right", portText);
  const controls = node("div", "flex shrink-0 flex-wrap gap-2");
  if (container.openUrl) controls.append(action("Open port", { href: container.openUrl, external: true }));
  controls.append(action("Copy ID", { onClick: () => copyText(container.id, "Container ID") }));
  row.append(info, ports, controls);
  return row;
}

function serviceCard(service) {
  const card = node("article", "surface flex min-h-28 flex-col gap-4 p-4 transition hover:border-violet-300/25");
  const top = node("div", "flex items-start justify-between gap-3");
  const info = node("div", "min-w-0 flex-1");
  info.append(
    routeLink(service.name, servicePath(service), "block truncate text-sm font-bold text-white hover:text-violet-200 focus:outline-none focus:ring-2 focus:ring-violet-400/60"),
    node("p", "mt-1 truncate text-xs capitalize text-slate-500", service.category),
  );
  top.append(info, node("span", "badge shrink-0", service.category));
  const controls = node("div", "mt-auto flex flex-wrap gap-2");
  controls.append(
    routeLink("Details", servicePath(service)),
    action("Copy start", { title: service.command, onClick: () => copyText(service.command, "Compose command") }),
  );
  card.append(top, controls);
  return card;
}

function collectionGrid(items, cardBuilder, emptyMessage, classes) {
  const grid = node("div", classes);
  if (!items.length) grid.append(emptyState(emptyMessage));
  else for (const item of items) grid.append(cardBuilder(item));
  return grid;
}

function relatedContainers(name) {
  const needle = name.toLocaleLowerCase().replaceAll("_", "-");
  return state.inventory.containers.filter((container) => {
    const candidate = container.name.toLocaleLowerCase().replaceAll("_", "-");
    return candidate === needle || candidate.includes(`-${needle}`) || candidate.includes(`${needle}-`);
  });
}

function statCard(label, value, href, tone = "text-white") {
  const card = routeLink("", href, "stat-card block transition hover:border-teal-300/30 focus:outline-none focus:ring-2 focus:ring-teal-400/60");
  card.append(node("p", "text-xs font-bold uppercase tracking-[0.16em] text-slate-500", label), node("p", `mt-2 text-3xl font-bold ${tone}`, String(value)));
  return card;
}

function renderHome() {
  setPage("Local workspace", "Dashboard", "Your apps, containers, and services at a glance.", { search: true, nav: "home" });
  const fragment = document.createDocumentFragment();
  const stats = node("div", "mb-10 grid grid-cols-2 gap-3 lg:grid-cols-4");
  stats.append(
    statCard("Apps", state.inventory.projects.length, "/apps", "text-teal-200"),
    statCard("Running", state.inventory.containers.length, "/containers", "text-blue-200"),
    statCard("Services", state.inventory.services.length, "/services", "text-violet-200"),
    statCard("Categories", new Set(state.inventory.services.map((service) => service.category)).size, "/services", "text-amber-200"),
  );
  fragment.append(stats);

  const apps = filtered(state.inventory.projects).slice(0, 6);
  const appSection = node("section", "mb-10");
  appSection.append(sectionHeader("Workspaces", "Apps", state.inventory.projects.length, "/apps"));
  appSection.append(collectionGrid(apps, projectCard, "No apps match your search.", "grid gap-4 md:grid-cols-2 xl:grid-cols-3"));
  fragment.append(appSection);

  const containers = filtered(state.inventory.containers).slice(0, 6);
  const containerSection = node("section", "mb-10");
  containerSection.append(sectionHeader("Podman", "Running containers", state.inventory.containers.length, "/containers"));
  const list = node("div", "surface divide-y divide-white/5 overflow-hidden");
  if (!containers.length) list.append(emptyState("No running containers match your search."));
  else for (const container of containers) list.append(containerRow(container));
  containerSection.append(list);
  fragment.append(containerSection);

  const services = filtered(state.inventory.services).slice(0, 8);
  const serviceSection = node("section");
  serviceSection.append(sectionHeader("Catalog", "Services", state.inventory.services.length, "/services"));
  serviceSection.append(collectionGrid(services, serviceCard, "No services match your search.", "grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"));
  fragment.append(serviceSection);
  elements.view.replaceChildren(fragment);
}

function renderApps() {
  const apps = filtered(state.inventory.projects);
  setPage("Workspaces", "Apps", "Browse every application workspace and open its local site or detail page.", { search: true, nav: "apps" });
  const content = node("section");
  content.append(sectionHeader("Available", "All apps", apps.length));
  content.append(collectionGrid(apps, projectCard, state.query ? "No apps match your search." : "No app workspaces found.", "grid gap-4 md:grid-cols-2 xl:grid-cols-3"));
  elements.view.replaceChildren(content);
}

function renderContainers() {
  const containers = filtered(state.inventory.containers);
  setPage("Podman", "Running containers", "Inspect the current runtime inventory and open published ports.", { search: true, nav: "containers" });
  const content = node("section");
  content.append(sectionHeader("Runtime", "Containers", containers.length));
  const list = node("div", "surface divide-y divide-white/5 overflow-hidden");
  if (!containers.length) list.append(emptyState(state.query ? "No containers match your search." : "No running containers found."));
  else for (const container of containers) list.append(containerRow(container));
  content.append(list);
  elements.view.replaceChildren(content);
}

function renderServices() {
  const services = filtered(state.inventory.services);
  setPage("Compose catalog", "Services", "Explore the DevArch service library by category and copy native startup commands.", { search: true, nav: "services" });
  const content = node("section");
  content.append(sectionHeader("Catalog", "All services", services.length));
  content.append(collectionGrid(services, serviceCard, state.query ? "No services match your search." : "No Compose services found.", "grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"));
  elements.view.replaceChildren(content);
}

function detailField(label, value, monospace = false) {
  const field = node("div", "stat-card min-w-0");
  field.append(node("dt", "text-xs font-bold uppercase tracking-[0.16em] text-slate-500", label));
  field.append(node("dd", `mt-2 break-words text-sm text-slate-200 ${monospace ? "font-mono" : ""}`, value));
  return field;
}

function renderAppDetail(name) {
  const project = state.inventory.projects.find((item) => item.name === name);
  if (!project) return renderNotFound("App not found", "This workspace is not present under apps/.", "/apps", "Back to apps");

  setPage("App workspace", project.name, `${project.kind} project in ${project.relativePath}`, { search: false, nav: "apps", back: ["/apps", "← All apps"] });
  const related = relatedContainers(project.name);
  const content = node("div", "grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]");
  const main = node("section", "surface p-5 sm:p-7");
  const heading = node("div", "flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between");
  const identity = node("div", "min-w-0");
  identity.append(node("p", "font-mono text-sm text-teal-300", project.url), node("p", "mt-2 text-sm leading-6 text-slate-400", "Open the local site, reveal the workspace in your editor, or copy its path."));
  heading.append(identity, node("span", "badge shrink-0", project.kind));
  const actions = node("div", "mt-6 flex flex-wrap gap-2");
  actions.append(
    action("Open site", { href: project.url, external: true, primary: true }),
    action("Open folder", { href: `vscode://file${encodeURI(project.path)}`, title: "Open with a VS Code-compatible editor" }),
    action("Copy path", { onClick: () => copyText(project.path, "Path") }),
  );
  const fields = node("dl", "mt-8 grid gap-3 sm:grid-cols-2");
  fields.append(detailField("Type", project.kind), detailField("Workspace", project.relativePath, true), detailField("Local domain", project.url, true), detailField("Related containers", String(related.length)));
  main.append(heading, actions, fields);

  const aside = node("aside", "surface p-5");
  aside.append(node("h2", "text-base font-bold text-white", "Related containers"));
  if (!related.length) aside.append(node("p", "mt-3 text-sm leading-6 text-slate-500", "No running container name currently matches this app."));
  else {
    const list = node("div", "mt-4 space-y-3");
    for (const container of related) {
      const item = node("div", "rounded-xl border border-white/10 bg-slate-950/50 p-3");
      item.append(node("p", "truncate text-sm font-semibold text-white", container.name), node("p", "mt-1 truncate text-xs text-slate-500", container.status || container.state));
      list.append(item);
    }
    aside.append(list);
  }
  content.append(main, aside);
  elements.view.replaceChildren(content);
}

function renderServiceDetail(category, name) {
  const service = state.inventory.services.find((item) => item.category === category && item.name === name);
  if (!service) return renderNotFound("Service not found", "This Compose service is not present in the catalog.", "/services", "Back to services");

  setPage("Compose service", service.name, `${service.category} service from the DevArch catalog.`, { search: false, nav: "services", back: ["/services", "← All services"] });
  const related = relatedContainers(service.name);
  const content = node("div", "grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]");
  const main = node("section", "surface p-5 sm:p-7");
  const top = node("div", "flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between");
  const copy = node("div");
  copy.append(node("p", "text-sm leading-6 text-slate-400", "Start this service with native Podman Compose from its catalog directory."));
  top.append(copy, node("span", "badge capitalize", service.category));
  const command = node("div", "mt-6 rounded-xl border border-white/10 bg-slate-950/80 p-4");
  command.append(node("p", "break-all font-mono text-sm text-violet-200", service.command));
  const actions = node("div", "mt-4 flex flex-wrap gap-2");
  actions.append(
    action("Copy start command", { primary: true, onClick: () => copyText(service.command, "Compose command") }),
    action("Copy path", { onClick: () => copyText(service.path, "Service path") }),
  );
  const fields = node("dl", "mt-8 grid gap-3 sm:grid-cols-2");
  fields.append(detailField("Service ID", service.id, true), detailField("Category", service.category), detailField("Catalog path", service.relativePath, true), detailField("Matching containers", String(related.length)));
  main.append(top, command, actions, fields);

  const aside = node("aside", "surface p-5");
  aside.append(node("h2", "text-base font-bold text-white", "Runtime match"));
  if (!related.length) aside.append(node("p", "mt-3 text-sm leading-6 text-slate-500", "No running container name currently matches this catalog service."));
  else {
    const list = node("div", "mt-4 space-y-3");
    for (const container of related) {
      const item = node("div", "rounded-xl border border-white/10 bg-slate-950/50 p-3");
      item.append(node("p", "truncate text-sm font-semibold text-white", container.name), node("p", "mt-1 truncate text-xs text-slate-500", container.status || container.state));
      list.append(item);
    }
    aside.append(list);
  }
  content.append(main, aside);
  elements.view.replaceChildren(content);
}

function renderNotFound(title, description, href = "/", label = "Back to dashboard") {
  setPage("Not found", title, description, { search: false });
  const card = node("div", "surface px-6 py-14 text-center");
  card.append(node("p", "text-sm text-slate-500", description), routeLink(label, href, "primary-action mt-6"));
  elements.view.replaceChildren(card);
}

function currentRoute() {
  const parts = window.location.pathname.split("/").filter(Boolean).map((part) => decodeURIComponent(part));
  if (!parts.length) return { name: "home" };
  if (parts[0] === "projects") {
    const canonical = parts.length === 1 ? "/apps" : `/apps/${encodeURIComponent(parts[1])}`;
    window.history.replaceState({}, "", canonical);
    return parts.length === 1 ? { name: "apps" } : { name: "app", app: parts[1] };
  }
  if (parts[0] === "apps" && parts.length === 1) return { name: "apps" };
  if (parts[0] === "apps" && parts.length === 2) return { name: "app", app: parts[1] };
  if (parts[0] === "containers" && parts.length === 1) return { name: "containers" };
  if (parts[0] === "services" && parts.length === 1) return { name: "services" };
  if (parts[0] === "services" && parts.length === 3) return { name: "service", category: parts[1], service: parts[2] };
  return { name: "not-found" };
}

function setPage(eyebrow, title, description, options = {}) {
  document.title = `${title} · DevArch`;
  elements.pageEyebrow.textContent = eyebrow;
  elements.pageTitle.textContent = title;
  elements.pageDescription.textContent = description;
  elements.searchPanel.classList.toggle("hidden", !options.search);
  elements.pageBack.classList.toggle("hidden", !options.back);
  if (options.back) {
    elements.pageBack.href = options.back[0];
    elements.pageBack.textContent = options.back[1];
  }
  document.querySelectorAll("[data-nav]").forEach((link) => {
    if (link.dataset.nav === options.nav) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  });
}

function render() {
  if (!state.loaded) return;
  const runtimeError = state.inventory.runtimeError;
  elements.runtimeAlert.textContent = runtimeError || "";
  elements.runtimeAlert.classList.toggle("hidden", !runtimeError);
  const route = currentRoute();
  if (route.name === "home") renderHome();
  else if (route.name === "apps") renderApps();
  else if (route.name === "app") renderAppDetail(route.app);
  else if (route.name === "containers") renderContainers();
  else if (route.name === "services") renderServices();
  else if (route.name === "service") renderServiceDetail(route.category, route.service);
  else renderNotFound("Page not found", "This DevArch page does not exist.");
}

function setMenu(open) {
  elements.mobileMenu.classList.toggle("hidden", !open);
  elements.menuOpenIcon.classList.toggle("hidden", open);
  elements.menuCloseIcon.classList.toggle("hidden", !open);
  elements.menuButton.setAttribute("aria-expanded", String(open));
  elements.menuButton.setAttribute("aria-label", open ? "Close navigation menu" : "Open navigation menu");
}

function navigate(href) {
  window.history.pushState({}, "", href);
  state.query = "";
  elements.search.value = "";
  setMenu(false);
  render();
  window.scrollTo({ top: 0, behavior: "auto" });
}

async function refreshInventory() {
  elements.refreshButton.disabled = true;
  elements.refreshStatus.textContent = "Refreshing…";
  try {
    const response = await fetch("/api/inventory", { cache: "no-store" });
    if (!response.ok) throw new Error(`Inventory request failed (${response.status})`);
    state.inventory = await response.json();
    state.loaded = true;
    render();
    const refreshed = new Date(state.inventory.generatedAt);
    elements.refreshStatus.textContent = `Refreshed ${refreshed.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
  } catch (error) {
    elements.refreshStatus.textContent = "Refresh failed";
    elements.runtimeAlert.textContent = error.message;
    elements.runtimeAlert.classList.remove("hidden");
  } finally {
    elements.refreshButton.disabled = false;
  }
}

document.addEventListener("click", (event) => {
  const link = event.target.closest("a[data-route]");
  if (!link || event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
  if (link.origin !== window.location.origin) return;
  event.preventDefault();
  navigate(`${link.pathname}${link.search}${link.hash}`);
});

elements.search.addEventListener("input", (event) => {
  state.query = event.target.value.trim().toLocaleLowerCase();
  render();
});
elements.refreshButton.addEventListener("click", refreshInventory);
elements.menuButton.addEventListener("click", () => setMenu(elements.menuButton.getAttribute("aria-expanded") !== "true"));
window.addEventListener("popstate", () => {
  state.query = "";
  elements.search.value = "";
  setMenu(false);
  render();
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") setMenu(false);
  if (event.key === "/" && !elements.searchPanel.classList.contains("hidden") && document.activeElement !== elements.search) {
    event.preventDefault();
    elements.search.focus();
  }
});

refreshInventory();
