"use strict";

// In-memory prototype only. Never persist credentials or contact model endpoints.
const $ = (selector) => document.querySelector(selector);
const make = (tag, className, text) => {
  const node = document.createElement(tag);
  node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
};
const catalog = [
  {
    id: "deepseek",
    name: "DeepSeek",
    mark: "D",
    url: "https://api.deepseek.com",
    protocol: "chat",
    models: [{ id: "deepseek-chat", kind: "chat", dimension: "" }],
    description: "对话与推理 · 不提供预设 Embedding",
  },
  {
    id: "openai",
    name: "OpenAI",
    mark: "O",
    url: "https://api.openai.com/v1",
    protocol: "chat",
    models: [
      { id: "gpt-4.1-mini", kind: "chat", dimension: "" },
      { id: "text-embedding-3-small", kind: "embedding", dimension: "1536" },
    ],
    description: "对话与 Embedding · 模型为示例",
  },
  {
    id: "anthropic",
    name: "Anthropic",
    mark: "A",
    url: "https://api.anthropic.com/v1",
    protocol: "anthropic",
    models: [{ id: "claude-sonnet-4-5", kind: "chat", dimension: "" }],
    description: "原生 Messages 协议 · 对话模型",
  },
  {
    id: "siliconflow",
    name: "硅基流动",
    mark: "S",
    url: "https://api.siliconflow.cn/v1",
    protocol: "chat",
    models: [{ id: "BAAI/bge-m3", kind: "embedding", dimension: "1024" }],
    description: "第三方供应商 · OpenAI 兼容接口",
  },
];
const protocols = [
  { key: "chat", label: "OpenAI Chat Completions" },
  { key: "responses", label: "OpenAI Responses" },
  { key: "anthropic", label: "Anthropic Messages" },
];
let providers = [
  {
    ...structuredClone(catalog[0]),
    auth: "key",
    credential: true,
    custom: false,
  },
  {
    ...structuredClone(catalog[3]),
    auth: "key",
    credential: false,
    custom: false,
  },
];
let defaults = { chat: "", embedding: "" };
let savedDefaults = { chat: "", embedding: "", embeddingSignature: "" };
let draft = null;
let editingID = null;
let credentialCleared = false;
let revision = 0;
let pendingConfirmation = null;
let toastTimer;

function notify(text) {
  clearTimeout(toastTimer);
  $("#toast").textContent = text;
  $("#toast").hidden = false;
  toastTimer = setTimeout(() => {
    $("#toast").hidden = true;
  }, 3500);
}

function closePickers() {
  document.querySelectorAll(".picker-menu").forEach((menu) => {
    menu.hidden = true;
  });
  document
    .querySelectorAll(".picker-trigger")
    .forEach((trigger) => trigger.setAttribute("aria-expanded", "false"));
}

function picker(root, labelID, options, value, change) {
  root.replaceChildren();
  const button = make("button", "picker-trigger");
  button.type = "button";
  button.setAttribute("aria-haspopup", "menu");
  button.setAttribute("aria-expanded", "false");
  button.setAttribute("aria-labelledby", labelID);
  const selected = options.find((option) => option.key === value);
  button.append(
    make("span", "", selected?.label || "选择模型"),
    make("span", "", "⌄"),
  );
  const menu = make("div", "picker-menu");
  menu.hidden = true;
  menu.id = `${root.id}-menu`;
  menu.setAttribute("role", "menu");
  menu.setAttribute("aria-labelledby", labelID);
  button.setAttribute("aria-controls", menu.id);
  if (!options.length)
    menu.append(
      make("p", "picker-empty", "暂无可选模型。请先添加对应用途的模型供应商。"),
    );
  options.forEach((option) => {
    const item = make("button", "picker-option");
    item.type = "button";
    item.setAttribute("role", "menuitemradio");
    item.setAttribute("aria-checked", String(value === option.key));
    item.append(
      make("span", "", option.label),
      make("span", "", value === option.key ? "✓" : ""),
    );
    item.addEventListener("click", () => {
      change(option.key);
      closePickers();
      root.querySelector("button")?.focus();
    });
    menu.append(item);
  });
  function openMenu(last = false) {
    closePickers();
    menu.hidden = false;
    button.setAttribute("aria-expanded", "true");
    const items = menu.querySelectorAll("button");
    (last
      ? items[items.length - 1]
      : menu.querySelector('[aria-checked="true"]') || items[0]
    )?.focus();
  }
  button.addEventListener("click", () => {
    const wasOpen = !menu.hidden;
    closePickers();
    if (!wasOpen) openMenu();
  });
  root.onkeydown = (event) => {
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      closePickers();
      button.focus();
      return;
    }
    if (event.key === "Tab") {
      closePickers();
      return;
    }
    if (!["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) return;
    event.preventDefault();
    if (menu.hidden) {
      openMenu(event.key === "ArrowUp");
      return;
    }
    const items = Array.from(menu.querySelectorAll("button"));
    const index = items.indexOf(document.activeElement);
    const next =
      event.key === "Home"
        ? 0
        : event.key === "End"
          ? items.length - 1
          : (index + (event.key === "ArrowDown" ? 1 : -1) + items.length) %
            items.length;
    items[next]?.focus();
  };
  root.append(button, menu);
}

function modelOptions(kind) {
  return providers.flatMap((provider) =>
    provider.models
      .filter((model) => model.kind === kind)
      .map((model) => ({
        key: `${provider.id}:${model.id}`,
        label: `${provider.name} / ${model.id}`,
      })),
  );
}
function getEmbeddingSignature(key) {
  for (const provider of providers) {
    const model = provider.models.find(
      (model) =>
        `${provider.id}:${model.id}` === key && model.kind === "embedding",
    );
    if (model)
      return JSON.stringify([
        provider.id,
        provider.url,
        model.id,
        model.dimension,
      ]);
  }
  return "";
}
function renderDefaults() {
  for (const kind of ["chat", "embedding"]) {
    if (!modelOptions(kind).some((option) => option.key === defaults[kind]))
      defaults[kind] = "";
    picker(
      $(`#${kind}-picker`),
      `${kind}-label`,
      modelOptions(kind),
      defaults[kind],
      (value) => {
        defaults[kind] = value;
        renderDefaults();
      },
    );
  }
  const signature = getEmbeddingSignature(defaults.embedding);
  $("#dimension-hint").textContent = signature
    ? `${JSON.parse(signature)[3]} 维 · 模型维度需通过实际调用验证`
    : "尚未选择；配置完成后才可执行向量化。";
  const same =
    defaults.chat === savedDefaults.chat &&
    defaults.embedding === savedDefaults.embedding &&
    signature === savedDefaults.embeddingSignature;
  $("#default-state").textContent =
    savedDefaults.chat && same ? "默认模型已保存 · 演示" : "尚未保存默认模型";
  $("#default-state").className =
    `tag ${savedDefaults.chat && same ? "tag-tint" : "tag-warning"}`;
}
function confirm(title, message, actionLabel, action) {
  $("#confirm-title").textContent = title;
  $("#confirm-message").textContent = message;
  $("#confirm-action").textContent = actionLabel;
  pendingConfirmation = action;
  $("#confirm-dialog").showModal();
}
function renderProviders() {
  const list = $("#provider-list");
  list.replaceChildren();
  if (!providers.length)
    list.append(
      make(
        "div",
        "card empty-providers",
        "还没有模型供应商。添加一个连接，让知识整理与检索开始工作。",
      ),
    );
  providers.forEach((provider) => {
    const card = make("article", "provider-card");
    const top = make("div", "provider-top");
    const logo = make(
      "span",
      `provider-logo ${provider.custom ? "custom" : ""}`,
      provider.custom ? "</>" : provider.mark,
    );
    const info = make("div", "grow");
    const title = make("div", "provider-title");
    title.append(
      make("h3", "", provider.name),
      make(
        "span",
        `tag ${provider.credential || provider.auth === "none" ? "tag-tint" : "tag-warning"}`,
        provider.auth === "none"
          ? "无认证"
          : provider.credential
            ? "密钥已配置"
            : "待配置密钥",
      ),
    );
    info.append(
      title,
      make(
        "p",
        "provider-sub",
        `${provider.id} · ${provider.custom ? "自定义 API" : "预设供应商"} · ${provider.url}`,
      ),
    );
    const tools = make("div", "provider-tools");
    const edit = make("button", "btn btn-secondary", "配置");
    edit.type = "button";
    edit.setAttribute("aria-label", `配置 ${provider.name}`);
    edit.onclick = () => openEditor(provider, true);
    const remove = make("button", "delete", "删除");
    remove.type = "button";
    remove.setAttribute("aria-label", `删除 ${provider.name}`);
    remove.onclick = () => {
      if (
        [savedDefaults.chat, savedDefaults.embedding].some((key) =>
          key.startsWith(`${provider.id}:`),
        )
      ) {
        notify("此供应商正被默认模型使用，请先切换并保存默认模型。");
        return;
      }
      confirm(
        `删除 ${provider.name}？`,
        "移除此供应商连接及演示凭据，不删除知识正文。使用它的未保存模型选择将被清除。",
        "删除供应商",
        () => {
          providers = providers.filter((entry) => entry.id !== provider.id);
          renderProviders();
          renderDefaults();
          notify("供应商已移除（演示）");
        },
      );
    };
    tools.append(edit, remove);
    top.append(logo, info, tools);
    const models = make("div", "provider-models");
    provider.models.forEach((model) => {
      const chip = make("span", "model-chip");
      chip.append(
        make("span", "", model.kind === "embedding" ? "向量" : "对话"),
        make("strong", "", model.id),
      );
      if (model.kind === "embedding")
        chip.append(make("span", "", `${model.dimension} 维`));
      models.append(chip);
    });
    card.append(top, models);
    list.append(card);
  });
}
function renderCatalog() {
  $("#catalog-list").replaceChildren();
  catalog.forEach((provider) => {
    const option = make("button", "catalog-option");
    option.type = "button";
    option.disabled = providers.some((entry) => entry.id === provider.id);
    const info = make("div", "grow");
    info.append(
      make("strong", "", provider.name),
      make("p", "", provider.description),
    );
    option.append(
      make("span", "provider-logo", provider.mark),
      info,
      make("span", "t-secondary", option.disabled ? "已添加" : "＋"),
    );
    option.onclick = () => {
      $("#catalog-dialog").close();
      openEditor(
        {
          ...structuredClone(provider),
          auth: "key",
          credential: false,
          custom: false,
        },
        false,
      );
    };
    $("#catalog-list").append(option);
  });
}
function result(text, state = "") {
  $("#form-result").hidden = false;
  $("#form-result").className = `form-result ${state}`;
  $("#form-result").textContent = text;
}
function resetTest() {
  revision++;
  $("#form-result").hidden = true;
}
function renderModels() {
  $("#draft-models").replaceChildren();
  draft.models.forEach((model, index) => {
    const row = make("div", "model-row");
    const idField = make("div", "field model-id-field");
    const idLabel = make("label", "field-label", "模型 ID");
    idLabel.htmlFor = `model-${index}`;
    const input = make("input", "input mono");
    input.id = `model-${index}`;
    input.required = true;
    input.placeholder = "例如 text-embedding-3-small";
    input.value = model.id;
    input.oninput = () => {
      model.id = input.value;
      resetTest();
    };
    idField.append(idLabel, input);
    const kindField = make("div", "field");
    const kindLabel = make("label", "field-label", "用途");
    kindLabel.id = `kind-label-${index}`;
    const kindPicker = make("div", "picker");
    kindPicker.id = `kind-${index}`;
    picker(
      kindPicker,
      kindLabel.id,
      [
        { key: "chat", label: "对话" },
        { key: "embedding", label: "Embedding" },
      ],
      model.kind,
      (kind) => {
        model.kind = kind;
        model.dimension = kind === "chat" ? "" : model.dimension;
        resetTest();
        renderModels();
      },
    );
    kindField.append(kindLabel, kindPicker);
    const dimensionField = make("div", "field");
    const dimensionLabel = make("label", "field-label", "向量维度");
    dimensionLabel.htmlFor = `dimension-${index}`;
    const dimension = make("input", "input mono");
    Object.assign(dimension, {
      id: `dimension-${index}`,
      type: "number",
      min: "1",
      max: "65536",
      step: "1",
      placeholder: "维度",
      value: model.dimension,
      disabled: model.kind !== "embedding",
      required: model.kind === "embedding",
    });
    dimension.oninput = () => {
      model.dimension = dimension.value;
      resetTest();
    };
    dimensionField.append(dimensionLabel, dimension);
    const remove = make("button", "remove-model", "×");
    remove.type = "button";
    remove.setAttribute("aria-label", `移除模型 ${index + 1}`);
    remove.onclick = () => {
      draft.models.splice(index, 1);
      resetTest();
      renderModels();
      $("#add-model").focus();
    };
    row.append(idField, kindField, dimensionField, remove);
    $("#draft-models").append(row);
  });
}
function renderEditorPickers() {
  picker(
    $("#protocol-picker"),
    "protocol-label",
    protocols,
    draft.protocol,
    (value) => {
      draft.protocol = value;
      resetTest();
      renderEditorPickers();
    },
  );
  picker(
    $("#auth-picker"),
    "auth-label",
    [
      { key: "key", label: "API Key" },
      { key: "none", label: "无认证（本地服务）" },
    ],
    draft.auth,
    (value) => {
      draft.auth = value;
      resetTest();
      renderEditorPickers();
    },
  );
  $("#key-field").hidden = draft.auth === "none";
}
function openEditor(provider = null, editing = false) {
  closePickers();
  revision++;
  editingID = editing ? provider.id : null;
  credentialCleared = false;
  draft = provider
    ? structuredClone(provider)
    : {
        id: "",
        name: "",
        url: "",
        protocol: "chat",
        auth: "key",
        credential: false,
        custom: true,
        models: [{ id: "", kind: "chat", dimension: "" }],
      };
  $("#editor-title").textContent = editing
    ? `配置 ${draft.name}`
    : draft.custom
      ? "添加自定义 API"
      : `连接 ${draft.name}`;
  $("#editor-eyebrow").textContent = draft.custom
    ? "自定义连接 / OpenAI 兼容网关与本地服务"
    : "供应商目录 / 预设连接";
  $("#provider-id").value = draft.id;
  $("#provider-id").disabled = editing || !draft.custom;
  $("#provider-name").value = draft.name;
  $("#base-url").value = draft.url;
  $("#api-key").value = "";
  $("#api-key").placeholder = draft.credential
    ? "已配置；留空保留，输入 demo 替换"
    : "仅输入 demo 或 demo-fail";
  $("#key-label").textContent = draft.credential
    ? "API Key · 已配置，不回显"
    : "API Key · 尚未配置";
  $("#form-result").hidden = true;
  $("#discovery-result").hidden = true;
  $("#discover-models").textContent = draft.custom
    ? "获取可用模型"
    : "查看预设模型";
  renderEditorPickers();
  renderModels();
  $("#editor-dialog").showModal();
}
function readDraft(validateModels = true, requireCredential = false) {
  const valid = validateModels
    ? $("#provider-form").reportValidity()
    : ["#provider-id", "#provider-name", "#base-url"].every(
        (selector) => $(selector).disabled || $(selector).reportValidity(),
      );
  if (!valid) return false;
  draft.id = $("#provider-id").value.trim();
  draft.name = $("#provider-name").value.trim();
  draft.url = $("#base-url").value.trim();
  const url = new URL(draft.url);
  if (
    !["https:", "http:"].includes(url.protocol) ||
    url.username ||
    url.password ||
    url.search ||
    url.hash
  ) {
    result("请填写不含凭据、查询参数或片段的 HTTP(S) 基础地址。", "error");
    return false;
  }
  if (
    /\/(chat\/completions|embeddings|responses|messages)\/?$/.test(url.pathname)
  ) {
    result("请输入基础地址，不要包含具体模型接口路径。", "error");
    return false;
  }
  if (!draft.name) {
    result("请填写显示名称。", "error");
    return false;
  }
  if (
    providers.some(
      (provider) => provider.id === draft.id && provider.id !== editingID,
    )
  ) {
    result("Provider ID 已存在，请换一个唯一标识。", "error");
    return false;
  }
  const key = $("#api-key").value;
  if (key && !["demo", "demo-fail"].includes(key)) {
    result(
      "这是设计预览，请勿输入真实密钥。仅使用 demo 或 demo-fail。",
      "error",
    );
    return false;
  }
  if (
    requireCredential &&
    draft.auth === "key" &&
    !key &&
    (!draft.credential || credentialCleared)
  ) {
    result("请填写演示密钥 demo；或者选择无认证的本地服务。", "error");
    return false;
  }
  if (validateModels) {
    const ids = draft.models.map((model) => model.id.trim());
    if (
      !ids.length ||
      ids.some((id) => !id) ||
      new Set(ids).size !== ids.length
    ) {
      result("至少添加一个模型，模型 ID 不能为空或重复。", "error");
      return false;
    }
    if (
      draft.protocol === "anthropic" &&
      draft.models.some((model) => model.kind === "embedding")
    ) {
      result(
        "此原型的 Anthropic 连接仅支持对话，请另加 OpenAI 兼容 Embedding 供应商。",
        "error",
      );
      return false;
    }
    draft.models.forEach((model) => {
      model.id = model.id.trim();
    });
  }
  return true;
}
function saveDraft() {
  draft.credential =
    draft.auth === "key" &&
    (!!$("#api-key").value || (draft.credential && !credentialCleared));
  const previous = providers.findIndex((provider) => provider.id === editingID);
  if (previous < 0) providers.push(structuredClone(draft));
  else providers[previous] = structuredClone(draft);
  $("#api-key").value = "";
  $("#editor-dialog").close();
  renderProviders();
  renderDefaults();
  notify("供应商配置已更新（仅当前预览，刷新还原）");
}
$("#defaults-form").onsubmit = (event) => {
  event.preventDefault();
  if (!defaults.chat || !defaults.embedding) {
    notify("请分别选择对话模型和 Embedding 模型。");
    return;
  }
  const selectedProviders = providers.filter((provider) =>
    [defaults.chat, defaults.embedding].some((key) =>
      key.startsWith(`${provider.id}:`),
    ),
  );
  if (
    selectedProviders.some(
      (provider) => provider.auth === "key" && !provider.credential,
    )
  ) {
    notify("选中的模型供应商还未配置密钥，请先完成连接配置。");
    return;
  }
  const save = () => {
    savedDefaults = {
      ...defaults,
      embeddingSignature: getEmbeddingSignature(defaults.embedding),
    };
    renderDefaults();
    notify("默认模型已保存（演示），后续新任务使用此配置。");
  };
  if (
    savedDefaults.embeddingSignature &&
    savedDefaults.embeddingSignature !==
      getEmbeddingSignature(defaults.embedding)
  )
    confirm(
      "切换向量化模型？",
      "新旧模型的向量不能混用，即使维度相同。保存后需要重建关联索引；此预览不会执行重建，原始知识正文保持不变。",
      "确认切换",
      save,
    );
  else save();
};
$("#provider-form").onsubmit = (event) => {
  event.preventDefault();
  if (!readDraft()) return;
  const old = providers.find((provider) => provider.id === editingID);
  if (
    old &&
    savedDefaults.chat.startsWith(`${old.id}:`) &&
    !draft.models.some(
      (model) =>
        model.kind === "chat" &&
        `${draft.id}:${model.id}` === savedDefaults.chat,
    )
  ) {
    result("此对话模型正被默认值使用，请先切换默认模型再移除它。", "error");
    return;
  }
  if (old && savedDefaults.embedding.startsWith(`${old.id}:`)) {
    const nextModel = draft.models.find(
      (model) =>
        `${draft.id}:${model.id}` === savedDefaults.embedding &&
        model.kind === "embedding",
    );
    if (!nextModel) {
      result(
        "此模型被已保存的 Embedding 默认值使用，请先切换默认模型再移除它。",
        "error",
      );
      return;
    }
    const oldModel = old.models.find(
      (model) => `${old.id}:${model.id}` === savedDefaults.embedding,
    );
    if (old.url !== draft.url || oldModel.dimension !== nextModel.dimension) {
      confirm(
        "更新当前向量化配置？",
        "修改当前 Embedding 地址或维度会使已有索引需要重建。保存仅更新配置，不会自动重建向量。",
        "保存配置",
        () => {
          saveDraft();
          savedDefaults.embeddingSignature = getEmbeddingSignature(
            savedDefaults.embedding,
          );
          renderDefaults();
          notify("配置已保存（演示）；已有向量需要重建。");
        },
      );
      return;
    }
  }
  saveDraft();
};
$("#test-provider").onclick = () => {
  if (!readDraft(true, true)) return;
  const currentRevision = revision;
  const failed = $("#api-key").value === "demo-fail";
  result("正在验证连接与模型能力…（模拟，不发送请求）");
  $("#test-provider").disabled = true;
  setTimeout(() => {
    $("#test-provider").disabled = false;
    if (revision !== currentRevision || !$("#editor-dialog").open) return;
    result(
      failed
        ? "认证失败（演示）：请检查供应商凭据。可换成 demo 体验成功状态。"
        : `验证成功（演示）：${draft.models.filter((model) => model.kind === "chat").length} 个对话模型，${draft.models.filter((model) => model.kind === "embedding").length} 个 Embedding 模型。正式实现会分别检查调用与返回维度。`,
      failed ? "error" : "success",
    );
  }, 650);
};
$("#discover-models").onclick = () => {
  if (!draft.custom) {
    $("#discovery-result").hidden = false;
    $("#discovery-result").textContent =
      "已使用预设模型示例，不发送请求。你也可以手动补充模型。";
    return;
  }
  if (!readDraft(false, true)) return;
  const region = $("#discovery-result");
  region.hidden = false;
  region.replaceChildren(
    make(
      "p",
      "",
      "模拟 /models 候选。选择只加入草稿；用途和维度不会自动推断。若服务不支持发现，可直接手动添加。",
    ),
  );
  const actions = make("div", "actions");
  ["gateway-chat", "gateway-embedding"].forEach((id) => {
    const button = make("button", "btn btn-secondary", `＋ ${id}`);
    button.type = "button";
    button.onclick = () => {
      if (draft.models.some((model) => model.id === id)) {
        result("此模型已经在草稿中。");
        return;
      }
      draft.models.push({ id, kind: "chat", dimension: "" });
      resetTest();
      renderModels();
    };
    actions.append(button);
  });
  region.append(actions);
};
$("#clear-key").onclick = () => {
  credentialCleared = true;
  draft.credential = false;
  $("#api-key").value = "";
  $("#key-label").textContent = "API Key · 待清除或替换";
  resetTest();
  result(
    "凭据已从草稿清除；取消不影响已保存配置。保存后将显示待配置，模型调用暂不可用。",
  );
};
$("#add-model").onclick = () => {
  draft.models.push({ id: "", kind: "chat", dimension: "" });
  resetTest();
  renderModels();
  $("#draft-models .model-row:last-child input")?.focus();
};
$("#provider-form").addEventListener("input", resetTest);
$("#add-catalog").onclick = () => {
  renderCatalog();
  $("#catalog-dialog").showModal();
};
$("#add-custom").onclick = () => openEditor();
$("#catalog-custom").onclick = () => {
  $("#catalog-dialog").close();
  openEditor();
};
$("#confirm-action").onclick = () => {
  const action = pendingConfirmation;
  pendingConfirmation = null;
  $("#confirm-dialog").close();
  action?.();
};
document.querySelectorAll("[data-close]").forEach((button) => {
  button.onclick = () => button.closest("dialog").close();
});
$("#editor-dialog").addEventListener("close", () => {
  revision++;
  $("#api-key").value = "";
  $("#test-provider").disabled = false;
  closePickers();
});
document.addEventListener("click", (event) => {
  if (!event.target.closest(".picker")) closePickers();
});
function route() {
  const view = ["models", "services", "notes"].includes(location.hash.slice(1))
    ? location.hash.slice(1)
    : "models";
  document.querySelectorAll("main > section").forEach((section) => {
    section.hidden = section.id !== `view-${view}`;
  });
  document.querySelectorAll("[data-view]").forEach((link) => {
    if (link.dataset.view === view) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  });
}
$("#check-services").onclick = () => {
  document.querySelectorAll(".service-state").forEach((state) => {
    state.textContent = "检测中 · 演示";
    state.className = "tag tag-gray service-state";
  });
  $("#check-services").disabled = true;
  setTimeout(() => {
    document.querySelectorAll(".service-state").forEach((state) => {
      state.textContent = "可连接 · 演示";
      state.className = "tag tag-tint service-state";
    });
    $("#check-services").disabled = false;
    notify("服务检测完成（模拟），未访问本机服务。");
  }, 650);
};
window.addEventListener("hashchange", route);
renderProviders();
renderDefaults();
route();
