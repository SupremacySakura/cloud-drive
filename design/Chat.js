"use strict";

// Prototype only: no credentials, persistence, model calls, or tools.
const $ = (selector) => document.querySelector(selector);
const models = [
  { name: "DeepSeek", id: "deepseek-chat", default: true },
  { name: "我的模型网关", id: "chat-model", default: false },
];
let selected = 0;
let messages = [];
let running = false;
let missing = false;
let timer = null;
let toastTimer = null;
let pendingAction = null;
let returnFocus = null;
let callState = "尚未验证调用";
const reply =
  "你好！我可以帮助你解释概念、整理思路和进行文字交流。\n\n这是普通对话，不会读取知序中的文件或知识库，也不会调用任何工具。\n\n当前页面是设计预览，这段回复是模拟内容，不代表真实模型接入已成功。";
function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}
function notify(text) {
  clearTimeout(toastTimer);
  $("#toast").textContent = text;
  $("#toast").hidden = false;
  toastTimer = setTimeout(() => {
    $("#toast").hidden = true;
  }, 2600);
}
function modelName() {
  return `${models[selected].name} / ${models[selected].id}`;
}
function updateControls() {
  $("#send").disabled = running || missing || !$("#prompt").value.trim();
  $("#send").hidden = running;
  $("#stop").hidden = !running;
  $("#prompt").disabled = missing || running;
  $("#model-trigger").disabled = running || missing;
  $("#count").textContent = `${$("#prompt").value.length} / 8000`;
  $("#model-label").textContent = missing ? "尚未配置对话模型" : modelName();
  $("#connection-status").textContent = missing
    ? "请前往设置完成配置"
    : `${models[selected].default ? "使用默认模型" : "仅本次会话"} · ${callState}`;
  $("#live-status").textContent = running
    ? "正在生成 · 可停止，不影响已有消息"
    : "";
  $("#welcome").hidden = missing || messages.length > 0;
  $("#unconfigured").hidden = !missing;
  $("#messages").hidden = missing || messages.length === 0;
}
function scrollToEnd() {
  $("#transcript").scrollTop = $("#transcript").scrollHeight;
}
function render() {
  $("#messages").replaceChildren();
  messages.forEach((message, index) => {
    const article = element("article", `message ${message.role}`);
    const heading = element("div", "message-heading");
    heading.append(
      element("strong", "", message.role === "user" ? "你" : message.model),
    );
    if (message.role === "assistant")
      heading.append(element("span", "", "模拟回复"));
    const body = element(
      "div",
      "message-body",
      message.text ||
        (message.status === "generating" ? "正在等待首段回复…" : ""),
    );
    article.append(heading, body);
    if (message.role === "assistant") {
      const metadata = element("div", "message-status");
      const label = {
        generating: "正在生成…",
        completed: "回复完成 · 演示",
        stopped: "已停止 · 部分内容",
        failed: "未完成",
        interrupted: "连接中断 · 部分内容",
      };
      metadata.append(element("span", "", label[message.status]));
      if (message.text && message.status !== "generating") {
        const copy = element("button", "", "复制回复");
        copy.type = "button";
        copy.onclick = async () => {
          try {
            await navigator.clipboard.writeText(message.text);
            notify("回复已复制");
          } catch {
            notify("无法访问剪贴板，请手动选中文字复制。");
          }
        };
        metadata.append(copy);
      }
      article.append(metadata);
      if (message.error) {
        const box = element("div", "error-box");
        box.setAttribute("role", "alert");
        box.append(
          element("strong", "", message.error.title),
          element("p", "", message.error.detail),
        );
        const retry = element("button", "", "重试这条消息");
        retry.type = "button";
        retry.disabled = running;
        retry.onclick = () => {
          messages.splice(index);
          callState = "尚未验证调用";
          startReply();
        };
        const settings = element("a", "", "检查模型设置 ↗");
        settings.href = "./ModelSettings.html#models";
        box.append(retry, settings);
        article.append(box);
      }
    }
    $("#messages").append(article);
  });
  updateControls();
}
function startReply() {
  running = true;
  const message = {
    role: "assistant",
    model: modelName(),
    text: "",
    status: "generating",
  };
  messages.push(message);
  render();
  scrollToEnd();
  let offset = 0;
  timer = setInterval(() => {
    offset += 5;
    message.text = reply.slice(0, offset);
    if (offset >= reply.length) {
      clearInterval(timer);
      timer = null;
      running = false;
      message.status = "completed";
      callState = "已收到回复 · 演示";
    }
    const nearBottom =
      $("#transcript").scrollHeight -
        $("#transcript").scrollTop -
        $("#transcript").clientHeight <
      100;
    render();
    if (nearBottom) scrollToEnd();
    if (!running) {
      $("#live-status").textContent = "回复完成（模拟）";
      $("#prompt").focus();
    }
  }, 70);
}
function stopReply() {
  clearInterval(timer);
  timer = null;
  if (running) {
    const current = messages.at(-1);
    current.status = "stopped";
    if (!current.text) current.text = "生成已停止，未收到回复。";
  }
  running = false;
  render();
  $("#live-status").textContent = "已停止生成";
}
function reset() {
  clearInterval(timer);
  timer = null;
  running = false;
  messages = [];
  $("#prompt").value = "";
  callState = "尚未验证调用";
  render();
}
function confirmAction(description, action) {
  pendingAction = action;
  returnFocus = document.activeElement;
  $("#confirm-description").textContent = description;
  $("#confirm-dialog").showModal();
}
function closeConfirm() {
  $("#confirm-dialog").close();
  pendingAction = null;
  returnFocus?.focus();
}
$("#cancel-confirm").onclick = closeConfirm;
$("#confirm-dialog").addEventListener("cancel", (event) => {
  event.preventDefault();
  closeConfirm();
});
$("#accept-confirm").onclick = () => {
  const action = pendingAction;
  closeConfirm();
  action?.();
  $("#prompt").focus();
};
$("#new-chat").onclick = () => {
  if (messages.length || $("#prompt").value)
    confirmAction(
      "当前会话是临时的。新建将清空消息和输入草稿；正在生成的回复也会停止，无法恢复。",
      () => reset(),
    );
  else reset();
};
$("#composer").onsubmit = (event) => {
  event.preventDefault();
  const text = $("#prompt").value.trim();
  if (!text || running || missing) return;
  messages.push({ role: "user", text });
  $("#prompt").value = "";
  $("#prompt").style.height = "";
  startReply();
};
$("#prompt").oninput = () => {
  $("#prompt").style.height = "auto";
  $("#prompt").style.height = `${Math.min($("#prompt").scrollHeight, 180)}px`;
  updateControls();
};
$("#prompt").onkeydown = (event) => {
  if (event.key === "Enter" && !event.shiftKey && !event.isComposing) {
    event.preventDefault();
    $("#composer").requestSubmit();
  }
};
$("#stop").onclick = stopReply;
document.querySelectorAll("[data-prompt]").forEach((button) => {
  button.onclick = () => {
    $("#prompt").value = button.dataset.prompt;
    updateControls();
    $("#prompt").focus();
  };
});
function closeMenu(focus = false) {
  $("#model-menu").hidden = true;
  $("#model-trigger").setAttribute("aria-expanded", "false");
  if (focus) $("#model-trigger").focus();
}
function buildMenu() {
  $("#model-menu").replaceChildren();
  models.forEach((model, index) => {
    const button = element(
      "button",
      "",
      `${model.name} / ${model.id}${model.default ? " · 默认" : ""}`,
    );
    button.type = "button";
    button.setAttribute("role", "menuitemradio");
    button.setAttribute("aria-checked", String(index === selected));
    button.onclick = () => {
      closeMenu(true);
      if (selected === index) return;
      const choose = () => {
        selected = index;
        reset();
        buildMenu();
        notify("已切换本次会话模型，不修改默认设置。");
      };
      if (messages.length || $("#prompt").value)
        confirmAction(
          "切换模型将开始新对话并清空当前消息与草稿，避免混用上下文。不会修改设置中的默认模型。",
          choose,
        );
      else choose();
    };
    $("#model-menu").append(button);
  });
}
$("#model-trigger").onclick = () => {
  const menu = $("#model-menu");
  if (!menu.hidden) {
    closeMenu();
    return;
  }
  menu.hidden = false;
  $("#model-trigger").setAttribute("aria-expanded", "true");
  menu.querySelector('[aria-checked="true"]').focus();
};
$(".picker").addEventListener("keydown", (event) => {
  if ($("#model-menu").hidden) return;
  if (event.key === "Escape") {
    event.preventDefault();
    event.stopPropagation();
    closeMenu(true);
  }
  if (event.key === "Tab") closeMenu();
  if (["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) {
    event.preventDefault();
    const buttons = [...$("#model-menu").querySelectorAll("button")];
    let index = buttons.indexOf(document.activeElement);
    if (event.key === "Home") index = 0;
    else if (event.key === "End") index = buttons.length - 1;
    else
      index =
        (index + (event.key === "ArrowDown" ? 1 : -1) + buttons.length) %
        buttons.length;
    buttons[index].focus();
  }
});
document.addEventListener("click", (event) => {
  if (!$(".picker").contains(event.target)) closeMenu();
});
const errors = {
  auth: {
    title: "供应商拒绝了认证（模拟）",
    detail:
      "请检查 API Key、基础地址和模型 ID。这条消息未自动重发，已有对话仍保留。",
  },
  timeout: {
    title: "等待模型回复超时（模拟）",
    detail:
      "模型服务未在限定时间内响应。你可以检查服务后手动重试；重试可能再次产生调用费用。",
  },
  interrupted: {
    title: "回复连接中断（模拟）",
    detail:
      "已收到的文字保留，但回复不完整。重试会重新生成本条回复，不重复添加用户消息。",
  },
};
function scene(name) {
  closeMenu();
  missing = name === "missing";
  reset();
  if (name === "example") {
    messages = [
      { role: "user", text: "我的项目叫知序，请先记住这个名字。" },
      {
        role: "assistant",
        model: modelName(),
        text: "好的，在本次会话中，我会记住你的项目叫「知序」。",
        status: "completed",
      },
      { role: "user", text: "那我的项目叫什么？" },
      {
        role: "assistant",
        model: modelName(),
        text: "你的项目叫「知序」。\n\n这是多轮消息布局演示，并未调用模型，也不代表真实上下文传递已验证。",
        status: "completed",
      },
    ];
    callState = "已收到回复 · 演示";
  } else if (errors[name]) {
    messages = [
      { role: "user", text: "你好，请用一句话介绍你能做什么。" },
      {
        role: "assistant",
        model: modelName(),
        text:
          name === "interrupted"
            ? "你好！我可以帮你解释概念、整理思路，以及"
            : "",
        status: name === "interrupted" ? "interrupted" : "failed",
        error: errors[name],
      },
    ];
    callState = "本次调用未完成 · 演示";
  }
  render();
  scrollToEnd();
}
document.querySelectorAll("[data-scene]").forEach((button) => {
  button.onclick = () => scene(button.dataset.scene);
});
window.addEventListener("pagehide", () => {
  clearInterval(timer);
  clearTimeout(toastTimer);
});
buildMenu();
scene(location.hash.slice(1) || "ready");
window.addEventListener("hashchange", () =>
  scene(location.hash.slice(1) || "ready"),
);
