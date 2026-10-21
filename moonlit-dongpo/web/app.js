(() => {
  const opening = '明月几时有，把酒问青天。\n\n欢迎来到「月下东坡」。在这里，我们可以一起读懂苏轼的《水调歌头》：从词句、意象与音韵，到中秋之夜的思念与旷达。你想先从哪一句开始？';
  const messagesEl = document.getElementById('messages');
  const input = document.getElementById('chatInput');
  const send = document.getElementById('sendButton');
  const reset = document.getElementById('resetChat');
  const count = document.getElementById('charCount');
  const sessionId = createSessionId();
  let history = [];
  let busy = false;
  function createSessionId() { if (globalThis.crypto?.randomUUID) return crypto.randomUUID(); const b = new Uint8Array(20); crypto.getRandomValues(b); return Array.from(b, x => x.toString(16).padStart(2,'0')).join(''); }
  function textLength(s) { return [...s].length; }
  function createMessage(role, content, { error = false } = {}) {
    const row = document.createElement('div'); row.className = `message-row ${role === 'user' ? 'from-user' : 'from-bot'}`;
    const avatar = document.createElement('div'); avatar.className = 'msg-avatar'; avatar.textContent = role === 'user' ? '我' : '蘇';
    const body = document.createElement('div'); body.className = 'message-body';
    const author = document.createElement('span'); author.className = 'message-author'; author.textContent = role === 'user' ? '访客' : '东坡诗语';
    const bubble = document.createElement('div'); bubble.className = `message-bubble${error ? ' error' : ''}`; bubble.textContent = content;
    body.append(author,bubble); row.append(avatar,body); messagesEl.append(row);
    messagesEl.scrollTop = messagesEl.scrollHeight; return row;
  }
  function start() { history = [{ role:'assistant', content: opening }]; messagesEl.replaceChildren(); createMessage('assistant', opening); }
  function setBusy(value) { busy=value; send.disabled=value; input.disabled=value; document.querySelectorAll('.suggestion').forEach(b => b.disabled=value); send.textContent=value?'···':'➤'; }
  function syncCount() { count.textContent=`${textLength(input.value)} / 300`; input.style.height='auto'; input.style.height=Math.min(input.scrollHeight, 150)+'px'; }
  async function ask(question) {
    const text=(question ?? input.value).trim(); if (busy || !text || textLength(text)>300) return;
    input.value=''; syncCount(); createMessage('user',text); history.push({role:'user',content:text});
    const requestHistory=history.slice(-12); setBusy(true);
    const typing=createMessage('assistant','月光流转，正在寻句……'); typing.classList.add('typing');
    try {
      const response = await fetch('/api/chat', {method:'POST', headers:{'Content-Type':'application/json','X-Session-Id':sessionId},body:JSON.stringify({messages:requestHistory})});
      const result = await response.json(); typing.remove();
      if (!response.ok) { history.pop(); createMessage('assistant', result.reply || '暂时无法回应，请稍后再试。', {error:true}); return; }
      const reply = result.reply || '此刻无言，不妨再问一次。'; history.push({role:'assistant',content:reply}); createMessage('assistant',reply);
    } catch (_) { typing.remove(); history.pop(); createMessage('assistant','连接暂时中断，请检查网络后重试。',{error:true}); }
    finally { setBusy(false); input.focus(); }
  }
  send.addEventListener('click',() => ask());
  input.addEventListener('input',syncCount);
  input.addEventListener('keydown',e=>{if(e.key==='Enter'&&!e.shiftKey){e.preventDefault();ask();}});
  reset.addEventListener('click',()=>{if(!busy){start();input.focus();}});
  document.querySelectorAll('[data-question]').forEach(b=>b.addEventListener('click',()=>{document.getElementById('dialogue').scrollIntoView({behavior:'smooth'});ask(b.dataset.question);}));
  start();syncCount();
})();
