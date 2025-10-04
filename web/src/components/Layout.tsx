import React from 'react';
import { NavLink, Link, useNavigate } from 'react-router-dom';
import { clearToken, getToken, getUser } from '../lib/auth';

const navLinkClass = ({isActive}:{isActive:boolean}) =>
  'block px-3 py-2 rounded-md text-sm font-medium transition-colors ' +
  (isActive ? 'bg-brand-600 text-white shadow-sm' : 'text-slate-300 hover:text-white hover:bg-slate-700');

const Layout: React.FC<React.PropsWithChildren> = ({children}) => {
  const nav = useNavigate();
  const token = getToken();
  const user = getUser();
  const logout = () => { clearToken(); nav('/login'); };
  if(!token) return (
    <div className="min-h-screen flex items-center justify-center bg-slate-100 p-6">
      <div className="card w-full max-w-sm">
        <div className="card-body space-y-4">
          <p className="text-sm text-slate-600">Не авторизован.</p>
          <Link to="/login" className="btn w-full justify-center">Войти</Link>
        </div>
      </div>
    </div>
  );
  return (
    <div className="min-h-screen flex bg-slate-100">
      <aside className="w-60 bg-slate-900 text-slate-100 flex flex-col">
        <div className="h-14 flex items-center px-4 border-b border-slate-800">
          <span className="font-semibold tracking-wide">OSPAB</span>
        </div>
        <nav className="flex-1 px-3 py-3 space-y-1">
          <NavLink to="/" className={navLinkClass} end>Dashboard</NavLink>
          <NavLink to="/servers" className={navLinkClass}>Servers</NavLink>
          <NavLink to="/instances" className={navLinkClass}>Instances</NavLink>
          <NavLink to="/profile" className={navLinkClass}>Profile</NavLink>
        </nav>
        <div className="p-4 border-t border-slate-800 text-xs space-y-2">
          <div className="font-medium truncate">{user?.username}</div>
          <button onClick={logout} className="btn btn-secondary w-full justify-center !bg-slate-800 !text-slate-200 hover:!bg-slate-700">Выйти</button>
        </div>
      </aside>
      <main className="flex-1 p-6 overflow-auto">
        {children}
      </main>
    </div>
  );
};

export default Layout;
