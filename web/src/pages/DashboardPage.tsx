import React from 'react';
import { getUser } from '../lib/auth';

const DashboardPage: React.FC = () => {
  const user = getUser();
  const [meta,setMeta] = React.useState<{status:string;version:string} | null>(null);
  const [loading,setLoading] = React.useState(true);
  const [error,setError] = React.useState('');

  React.useEffect(()=>{
    const load = async () => {
      setLoading(true); setError('');
      try {
        const r = await fetch('/api/status',{headers:{'Authorization':`Bearer ${localStorage.getItem('ospab_token')}`}});
        if(!r.ok) throw new Error('Не удалось получить статус');
        const d = await r.json();
        setMeta({status:d.status,version:d.version});
      } catch(e:any){ setError(e.message); }
      finally { setLoading(false); }
    };
    load();
  },[]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-slate-800">Обзор</h1>
      </div>
      <div className="grid gap-5 md:grid-cols-3">
        <div className="card">
          <div className="card-body">
            <div className="text-xs uppercase tracking-wide text-slate-500 mb-1">API статус</div>
            {loading? <div className="text-sm text-slate-400">Загрузка…</div> : error? <div className="text-sm text-red-600">{error}</div> : (
              <div className="flex items-baseline gap-2">
                <span className={meta?.status==='ok' ? 'text-green-600 font-medium' : 'text-red-600 font-medium'}>{meta?.status}</span>
                <span className="text-xs text-slate-400">v{meta?.version}</span>
              </div>
            )}
          </div>
        </div>
        <div className="card">
          <div className="card-body">
            <div className="text-xs uppercase tracking-wide text-slate-500 mb-1">Пользователь</div>
            <div className="text-sm font-medium">{user?.username}</div>
            <div className="text-xs text-slate-500 break-all">{user?.email}</div>
          </div>
        </div>
        <div className="card">
          <div className="card-body">
            <div className="text-xs uppercase tracking-wide text-slate-500 mb-1">Быстрые действия</div>
            <div className="flex flex-wrap gap-2">
              <a href="/servers" className="btn btn-secondary text-xs py-1 px-3">Серверы</a>
              <a href="/instances" className="btn btn-secondary text-xs py-1 px-3">Инстансы</a>
              <a href="/profile" className="btn btn-secondary text-xs py-1 px-3">Профиль</a>
            </div>
          </div>
        </div>
      </div>
      <div className="grid gap-5 md:grid-cols-2">
        <div className="card">
          <div className="card-header"><span className="font-medium">Последние события</span></div>
          <div className="card-body text-sm text-slate-500">Пока нет данных (заглушка)</div>
        </div>
        <div className="card">
          <div className="card-header"><span className="font-medium">Статистика ресурсов</span></div>
          <div className="card-body text-sm text-slate-500 space-y-2">
            <div className="flex items-center justify-between"><span>CPU</span><span className="text-slate-400">—</span></div>
            <div className="flex items-center justify-between"><span>RAM</span><span className="text-slate-400">—</span></div>
            <div className="flex items-center justify-between"><span>Disk</span><span className="text-slate-400">—</span></div>
            <div className="flex items-center justify-between"><span>Network</span><span className="text-slate-400">—</span></div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
