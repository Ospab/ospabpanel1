import React from 'react';

interface ServerItem {
  id:number; name:string; host:string; port:number; type:string; is_active:boolean; username:string; created_at?:string; updated_at?:string;
}

type CreateForm = { name:string; host:string; port:number|string; type:string; username:string; password:string };
type EditForm = { name:string; host:string; port:number|string; username:string; password:string };

const ServersPage: React.FC = () => {
  const token = localStorage.getItem('ospab_token');
  const [items,setItems] = React.useState<ServerItem[]>([]);
  const [filtered,setFiltered] = React.useState<ServerItem[]>([]);
  const [loading,setLoading] = React.useState(false);
  const [error,setError] = React.useState('');
  const [creating,setCreating] = React.useState(false);
  const [form,setForm] = React.useState<CreateForm>({name:'',host:'',port:8006,type:'prx',username:'',password:''});
  const [editId,setEditId] = React.useState<number|null>(null);
  const [editForm,setEditForm] = React.useState<EditForm>({name:'',host:'',port:0,username:'',password:''});
  const [query,setQuery] = React.useState('');
  const [saving,setSaving] = React.useState(false);

  const authHeaders = React.useCallback(()=> ({'Authorization':`Bearer ${token}`,'Content-Type':'application/json'}),[token]);

  const applyFilter = React.useCallback((list:ServerItem[], q:string) => {
    if(!q.trim()) return list;
    const qq = q.toLowerCase();
    return list.filter(s => [s.name,s.host,String(s.port),s.type].some(v=>v.toLowerCase().includes(qq)));
  },[]);

  const load = async () => {
    setLoading(true); setError('');
    try {
      const res = await fetch(`/api/servers`, {headers: authHeaders()});
      if(!res.ok) throw new Error('Не удалось загрузить серверы');
      const data = await res.json();
      const list:ServerItem[] = Array.isArray(data)? data : (data.servers||[]);
      setItems(list);
      setFiltered(applyFilter(list, query));
    } catch(e:any){ setError(e.message || 'Ошибка загрузки'); } finally { setLoading(false); }
  };

  React.useEffect(()=>{ load(); /* eslint-disable-next-line react-hooks/exhaustive-deps */},[]);
  React.useEffect(()=>{ setFiltered(applyFilter(items, query)); },[items,query,applyFilter]);

  const handleCreate = async (e:React.FormEvent) => {
    e.preventDefault(); if(saving) return; setError(''); setSaving(true);
    try {
      const res = await fetch('/api/servers',{method:'POST',headers:authHeaders(),body:JSON.stringify({...form,port:Number(form.port)})});
      if(!res.ok){ const t = await res.text(); throw new Error(parseErrText(t)||'Ошибка сохранения'); }
      setCreating(false); setForm({name:'',host:'',port:8006,type:'prx',username:'',password:''});
      await load();
    } catch(e:any){ setError(e.message);} finally { setSaving(false); }
  };

  const startEdit = (s:ServerItem) => {
    setEditId(s.id);
    setEditForm({name:s.name,host:s.host,port:s.port,username:s.username,password:''});
  };
  const cancelEdit = () => { setEditId(null); };
  const handleUpdate = async (e:React.FormEvent) => {
    e.preventDefault(); if(editId==null) return; if(saving) return; setSaving(true); setError('');
    try {
      const payload:any = {};
      if(editForm.name) payload.name = editForm.name;
      if(editForm.host) payload.host = editForm.host;
      if(Number(editForm.port)>0) payload.port = Number(editForm.port);
      if(editForm.username) payload.username = editForm.username;
      if(editForm.password) payload.password = editForm.password; // пусто -> не менять
      const res = await fetch(`/api/servers/${editId}`,{method:'PUT',headers:authHeaders(),body:JSON.stringify(payload)});
      if(!res.ok){ const t = await res.text(); throw new Error(parseErrText(t)||'Ошибка обновления'); }
      setEditId(null); await load();
    } catch(e:any){ setError(e.message);} finally { setSaving(false); }
  };

  const handleDelete = async (id:number) => {
    if(!window.confirm('Удалить сервер? Он станет неактивным.')) return;
    try {
      const res = await fetch(`/api/servers/${id}`,{method:'DELETE',headers:authHeaders()});
      if(!res.ok) throw new Error('Не удалось удалить');
      setItems(prev => prev.filter(x=>x.id!==id));
    } catch(e:any){ setError(e.message); }
  };

  const parseErrText = (t:string) => {
    try { const j = JSON.parse(t); return j.error || j.message || t; } catch { return t; }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-xl font-semibold text-slate-800">Серверы</h1>
          <p className="text-xs text-slate-500 mt-1">Управление подключениями к гипервизорам. Тип нельзя изменить после создания.</p>
        </div>
        <div className="flex gap-2">
          <button onClick={load} disabled={loading} className="btn" title="Перезагрузить список">{loading? '...' : 'Обновить'}</button>
          <button onClick={()=>{ setCreating(v=>!v); setEditId(null); }} className="btn-secondary">{creating? 'Отмена':'Новый'}</button>
        </div>
      </div>
      <div className="flex flex-col md:flex-row gap-2 md:items-center">
        <div className="relative flex-1 max-w-xs">
          <input placeholder="Поиск..." value={query} onChange={e=>setQuery(e.target.value)} className="input pl-8" />
          <span className="absolute left-2 top-1.5 text-slate-400 text-sm">🔍</span>
        </div>
        <div className="text-xs text-slate-500">Всего: {filtered.length}</div>
      </div>
      {error && <div className="rounded-md bg-red-50 border border-red-200 px-4 py-2 text-sm text-red-700">{error}</div>}

      {creating && (
        <form onSubmit={handleCreate} className="card">
          <div className="card-header"><span className="font-medium">Добавить сервер</span></div>
          <div className="card-body grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="lbl">Имя</label>
              <input className="input" value={form.name} onChange={e=>setForm({...form,name:e.target.value})} required />
            </div>
            <div>
              <label className="lbl">Host</label>
              <input className="input" value={form.host} onChange={e=>setForm({...form,host:e.target.value})} required />
            </div>
            <div>
              <label className="lbl">Port</label>
              <input className="input" type="number" value={form.port} onChange={e=>setForm({...form,port:e.target.value})} required />
            </div>
            <div>
              <label className="lbl">Type</label>
              <select className="input" value={form.type} onChange={e=>setForm({...form,type:e.target.value})}>
                <option value="prx">Proxmox</option>
              </select>
            </div>
            <div>
              <label className="lbl">Username</label>
              <input className="input" value={form.username} onChange={e=>setForm({...form,username:e.target.value})} required />
            </div>
            <div>
              <label className="lbl">Password</label>
              <input className="input" type="password" value={form.password} onChange={e=>setForm({...form,password:e.target.value})} required />
            </div>
          </div>
          <div className="px-5 pb-4 flex justify-end gap-2">
            <button type="button" onClick={()=>setCreating(false)} className="btn-secondary">Отмена</button>
            <button className="btn" disabled={saving}>{saving? '...' : 'Сохранить'}</button>
          </div>
        </form>
      )}

      <div className="card">
        <div className="card-header"><span className="font-medium">Список серверов</span><span className="badge">{filtered.length}</span></div>
        <div className="card-body p-0">
          <div className="table-wrap">
            <table className="table">
              <thead className="thead">
                <tr>
                  <th className="th">ID</th>
                  <th className="th">Имя</th>
                  <th className="th">Хост</th>
                  <th className="th">Порт</th>
                  <th className="th">Тип</th>
                  <th className="th">Username</th>
                  <th className="th w-40">Создан</th>
                  <th className="th w-40">Обновлён</th>
                  <th className="th text-right">Действия</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map(s => (
                  <React.Fragment key={s.id}>
                    {editId === s.id ? (
                      <tr className="bg-amber-50">
                        <td className="td font-mono text-xs">{s.id}</td>
                        <td className="td">
                          <input className="input h-8" value={editForm.name} onChange={e=>setEditForm(f=>({...f,name:e.target.value}))} />
                        </td>
                        <td className="td">
                          <input className="input h-8" value={editForm.host} onChange={e=>setEditForm(f=>({...f,host:e.target.value}))} />
                        </td>
                        <td className="td w-24">
                          <input className="input h-8" type="number" value={editForm.port} onChange={e=>setEditForm(f=>({...f,port:e.target.value}))} />
                        </td>
                        <td className="td uppercase text-xs">{s.type}</td>
                        <td className="td">
                          <input className="input h-8" value={editForm.username} onChange={e=>setEditForm(f=>({...f,username:e.target.value}))} />
                        </td>
                        <td className="td text-xs text-slate-500">{s.created_at? new Date(s.created_at).toLocaleString(): '-'}</td>
                        <td className="td text-xs text-slate-500">{s.updated_at? new Date(s.updated_at).toLocaleString(): '-'}</td>
                        <td className="td">
                          <div className="flex justify-end gap-2">
                            <button className="btn-secondary h-8 px-2" onClick={cancelEdit} type="button">Отмена</button>
                            <button className="btn h-8 px-3" disabled={saving} onClick={handleUpdate}>{saving? '...' : 'Сохранить'}</button>
                          </div>
                          <div className="mt-2 text-[10px] text-slate-500">Пароль: <input placeholder="новый (не обяз.)" type="password" className="input h-7" value={editForm.password} onChange={e=>setEditForm(f=>({...f,password:e.target.value}))} /></div>
                        </td>
                      </tr>
                    ) : (
                      <tr className="hover:bg-slate-50">
                        <td className="td font-mono text-xs">{s.id}</td>
                        <td className="td">{s.name}</td>
                        <td className="td">{s.host}</td>
                        <td className="td">{s.port}</td>
                        <td className="td uppercase text-xs tracking-wide">{s.type}</td>
                        <td className="td text-xs">{s.username}</td>
                        <td className="td text-xs text-slate-500">{s.created_at? new Date(s.created_at).toLocaleDateString(): '-'}</td>
                        <td className="td text-xs text-slate-500">{s.updated_at? new Date(s.updated_at).toLocaleDateString(): '-'}</td>
                        <td className="td">
                          <div className="flex justify-end gap-1">
                            <button className="btn-secondary h-7 px-2" onClick={()=>startEdit(s)}>Изм</button>
                            <button className="btn-danger h-7 px-2" onClick={()=>handleDelete(s.id)}>Del</button>
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
                {filtered.length===0 && !loading && (
                  <tr>
                    <td colSpan={9} className="td text-center text-slate-500">Нет данных</td>
                  </tr>
                )}
                {loading && (
                  <tr>
                    <td colSpan={9} className="td text-center text-slate-400 text-sm">Загрузка...</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <div className="hidden">
        {/* utility classes short-hands */}
        <span className="lbl"></span>
      </div>
    </div>
  );
};

export default ServersPage;
