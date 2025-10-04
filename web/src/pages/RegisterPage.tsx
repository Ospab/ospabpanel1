import React from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { saveAuth, getToken } from '../lib/auth';

const API_BASE = (import.meta as any).env.VITE_API_URL || '';

const RegisterPage: React.FC = () => {
  const nav = useNavigate();
  const [loading,setLoading] = React.useState(false);
  const [error,setError] = React.useState('');
  const [showPassword,setShowPassword] = React.useState(false);

  React.useEffect(()=>{ if(getToken()) nav('/'); },[nav]);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault(); setError(''); setLoading(true);
    const fd = new FormData(e.target as HTMLFormElement);
    try {
      const res = await fetch(API_BASE + '/api/auth/register',{method:'POST',headers:{'Content-Type':'application/json'},body: JSON.stringify({username:fd.get('username'),email:fd.get('email'),password:fd.get('password')})});
      if(!res.ok){
        let msg = 'Ошибка регистрации';
        try { const j = await res.json(); if(j?.message) msg = j.message; } catch { const t = await res.text(); if(t) msg = t; }
        throw new Error(msg);
      }
      const data = await res.json();
      saveAuth(data.token, data.user);
      nav('/');
    } catch(err:any){ setError(err.message); } finally { setLoading(false); }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-100 to-slate-200 p-4">
      <div className="w-full max-w-lg">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold tracking-tight text-slate-800">Создать аккаунт</h1>
          <p className="text-sm text-slate-500 mt-2">Укажите данные для регистрации.</p>
        </div>
        <form onSubmit={submit} className="card">
          <div className="card-body space-y-5">
            <div className="grid md:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-slate-600 mb-1">Логин</label>
                <input name="username" className="input" required autoFocus />
              </div>
              <div>
                <label className="block text-xs font-medium text-slate-600 mb-1">Email</label>
                <input name="email" type="email" className="input" required />
              </div>
            </div>
            <div>
              <label className="flex items-center justify-between text-xs font-medium text-slate-600 mb-1">
                <span>Пароль</span>
                <button type="button" onClick={()=>setShowPassword(v=>!v)} className="text-brand-600 hover:underline text-[11px] font-normal">
                  {showPassword? 'Скрыть':'Показать'}
                </button>
              </label>
              <input name="password" type={showPassword? 'text':'password'} className="input" required />
            </div>
            {error && <div className="text-xs rounded-md bg-red-50 border border-red-200 px-3 py-2 text-red-600">{error}</div>}
            <button disabled={loading} className="btn w-full justify-center">{loading? '...' : 'Создать'}</button>
            <p className="text-xs text-slate-500 text-center">Уже есть аккаунт? <Link to="/login" className="text-brand-600 hover:underline">Войти</Link></p>
          </div>
        </form>
        <p className="mt-8 text-center text-[11px] text-slate-400">© {new Date().getFullYear()} OSPAB</p>
      </div>
    </div>
  );
};

export default RegisterPage;
