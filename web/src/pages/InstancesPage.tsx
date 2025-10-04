import React from 'react';

interface Server { id:number; name:string; type:string; }
interface Instance { id:string; name:string; type:string; status:string; }

const InstancesPage: React.FC = () => {
  const [servers,setServers] = React.useState<Server[]>([]);
  const [selected,setSelected] = React.useState<number|''>('');
  const [items,setItems] = React.useState<Instance[]>([]);
  const [loading,setLoading] = React.useState(false);
  const [error,setError] = React.useState('');
  const token = localStorage.getItem('ospab_token');

  React.useEffect(()=>{
    // загрузка серверов для выбора
    fetch('/api/servers',{headers:{'Authorization':`Bearer ${token}`}})
      .then(r=> r.ok? r.json(): Promise.reject())
      .then(d=> setServers(Array.isArray(d)? d : (d.servers||[])))
      .catch(()=>{});
  },[token]);

  const loadInstances = async () => {
    if(!selected) return; setLoading(true); setError('');
    try {
      const r = await fetch(`/api/servers/${selected}/instances`, {headers:{'Authorization':`Bearer ${token}`}});
      if(!r.ok) throw new Error('Не удалось загрузить');
      const d = await r.json();
      setItems(Array.isArray(d)? d : (d.instances||[]));
    } catch(e:any){ setError(e.message); } finally { setLoading(false); }
  };

  React.useEffect(()=>{ if(selected) loadInstances(); },[selected]);

  const action = async (id:string, act:string) => {
    try {
      await fetch(`/api/servers/${selected}/instances/${id}/${act}`, {method:'POST',headers:{'Authorization':`Bearer ${token}`}});
      loadInstances();
    } catch{}
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-slate-800">Инстансы</h1>
        <div className="flex items-center gap-2">
          <select className="input" value={selected} onChange={e=>setSelected(e.target.value? Number(e.target.value):'')}>
            <option value="">Выберите сервер</option>
            {servers.map(s=> <option key={s.id} value={s.id}>{s.name} ({s.type})</option>)}
          </select>
          <button onClick={loadInstances} disabled={!selected||loading} className="btn-secondary">Обновить</button>
        </div>
      </div>
      {!selected && <div className="text-sm text-slate-500">Сначала выберите сервер для просмотра инстансов.</div>}
      {error && <div className="text-sm text-red-600">{error}</div>}
      {selected && (
        <div className="card">
          <div className="card-header"><span className="font-medium">Список инстансов</span><span className="badge">{items.length}</span></div>
          <div className="card-body p-0">
            <div className="table-wrap">
              <table className="table">
                <thead className="thead">
                  <tr>
                    <th className="th">ID</th>
                    <th className="th">Имя</th>
                    <th className="th">Тип</th>
                    <th className="th">Статус</th>
                    <th className="th">Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {loading && (
                    <tr><td colSpan={5} className="td text-slate-500">Загрузка…</td></tr>
                  )}
                  {!loading && items.map(i=> (
                    <tr key={i.id} className="hover:bg-slate-50">
                      <td className="td font-mono text-xs">{i.id}</td>
                      <td className="td">{i.name}</td>
                      <td className="td uppercase text-xs">{i.type}</td>
                      <td className="td">{i.status}</td>
                      <td className="td space-x-2">
                        <button onClick={()=>action(i.id,'start')} className="text-xs text-brand-600 hover:underline">start</button>
                        <button onClick={()=>action(i.id,'stop')} className="text-xs text-brand-600 hover:underline">stop</button>
                        <button onClick={()=>action(i.id,'restart')} className="text-xs text-brand-600 hover:underline">restart</button>
                        <button onClick={()=>action(i.id,'status')} className="text-xs text-slate-500 hover:underline">status</button>
                      </td>
                    </tr>
                  ))}
                  {!loading && items.length===0 && (
                    <tr><td colSpan={5} className="td text-center text-slate-500">Нет данных</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default InstancesPage;
