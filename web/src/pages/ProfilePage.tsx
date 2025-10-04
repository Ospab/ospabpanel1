import React from 'react';
import { getUser } from '../lib/auth';

const ProfilePage: React.FC = () => {
  const user = getUser();
  const [editing,setEditing] = React.useState(false);
  const [form,setForm] = React.useState({username:user?.username||'', email:user?.email||''});

  const save = (e:React.FormEvent) => {
    e.preventDefault();
    // TODO: endpoint for update profile (заглушка)
    setEditing(false);
  };

  return (
    <div className="space-y-6 max-w-xl">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-slate-800">Профиль</h1>
        {!editing && <button className="btn-secondary" onClick={()=>setEditing(true)}>Редактировать</button>}
      </div>
      <div className="card">
        <div className="card-header"><span className="font-medium">Аккаунт</span></div>
        <div className="card-body">
          {user ? (
            editing ? (
              <form onSubmit={save} className="space-y-4">
                <div>
                  <label className="block text-xs font-medium text-slate-600 mb-1">Логин</label>
                  <input className="input" value={form.username} onChange={e=>setForm({...form,username:e.target.value})} />
                </div>
                <div>
                  <label className="block text-xs font-medium text-slate-600 mb-1">Email</label>
                  <input className="input" value={form.email} onChange={e=>setForm({...form,email:e.target.value})} />
                </div>
                <div className="flex justify-end gap-2 pt-2">
                  <button type="button" className="btn-secondary" onClick={()=>{setForm({username:user.username,email:user.email});setEditing(false);}}>Отмена</button>
                  <button className="btn">Сохранить</button>
                </div>
              </form>
            ) : (
              <div className="text-sm space-y-2">
                <div className="flex items-center justify-between"><span className="text-slate-500">ID</span><span className="font-mono">{user.id}</span></div>
                <div className="flex items-center justify-between"><span className="text-slate-500">Логин</span><span>{user.username}</span></div>
                <div className="flex items-center justify-between"><span className="text-slate-500">Email</span><span className="break-all">{user.email}</span></div>
              </div>
            )
          ) : <div className="text-sm text-slate-500">Нет данных пользователя</div>}
        </div>
      </div>
      <div className="card">
        <div className="card-header"><span className="font-medium">Безопасность</span></div>
        <div className="card-body text-sm text-slate-500 space-y-2">
          <p>Смена пароля пока не реализована.</p>
          <ul className="list-disc pl-5 space-y-1">
            <li>Минимум 8 символов</li>
            <li>Рекомендуется использовать менеджер паролей</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;
