import { useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { useStore } from '../store';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { setToken, setUser } = useStore();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await api.login(email, password);
      setToken(res.token);
      setUser(res.user);
    } catch (err: any) {
      setError(err.message || 'فشل تسجيل الدخول');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-900 via-purple-800 to-indigo-900">
      <div className="bg-white rounded-2xl shadow-2xl p-8 w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">ديوان</h1>
          <p className="text-gray-500">تسجيل الدخول إلى مساحة العمل</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          {error && (
            <div className="bg-red-50 text-red-600 p-3 rounded-lg text-sm">{error}</div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">البريد الإلكتروني</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none"
              placeholder="example@company.com"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">كلمة المرور</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent outline-none"
              placeholder="••••••••"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-purple-700 hover:bg-purple-800 text-white font-medium py-3 rounded-lg transition-colors disabled:opacity-50"
          >
            {loading ? 'جاري الدخول...' : 'تسجيل الدخول'}
          </button>
        </form>

        <p className="text-center mt-6 text-sm text-gray-500">
          ليس لديك حساب؟{' '}
          <Link to="/register" className="text-purple-700 hover:underline font-medium">
            إنشاء حساب جديد
          </Link>
        </p>
      </div>
    </div>
  );
}
