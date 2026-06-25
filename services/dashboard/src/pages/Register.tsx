import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { register } from '../api/gateway';

export default function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await register(email, password);
      navigate('/login');
    } catch (err: any) {
      setError(err.message || 'Registration failed');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <form onSubmit={handleRegister} className="bg-white p-8 rounded shadow-md w-96">
        <h2 className="text-2xl font-bold mb-6 text-center">Create Account</h2>
        {error && <p className="text-red-500 mb-4">{error}</p>}
        <input type="email" placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} className="w-full border p-2 mb-4 rounded" required />
        <input type="password" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} className="w-full border p-2 mb-6 rounded" required />
        <button type="submit" className="w-full bg-green-600 text-white py-2 rounded hover:bg-green-700">Register</button>
        <p className="mt-4 text-center text-sm">
          Already have an account? <Link to="/login" className="text-blue-600 underline">Sign in</Link>
        </p>
      </form>
    </div>
  );
}