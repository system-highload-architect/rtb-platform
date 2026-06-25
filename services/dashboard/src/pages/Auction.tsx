import { useState } from 'react';
import { Link } from 'react-router-dom';

export default function Auction() {
  const [deviceId, setDeviceId] = useState('test-device-1');
  const [ip, setIp] = useState('203.0.113.5');
  const [lat, setLat] = useState('55.7558');
  const [lng, setLng] = useState('37.6173');
  const [response, setResponse] = useState<any>(null);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setResponse(null);
    try {
      const res = await fetch('/rpc', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'auction.bid',
          params: {
            device_id: deviceId,
            ip,
            lat: parseFloat(lat),
            lng: parseFloat(lng),
          },
          id: 1,
        }),
      });
      const data = await res.json();
      setResponse(data);
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <div className="p-6 max-w-lg mx-auto">
      <nav className="mb-4">
        <Link to="/" className="mr-4 text-blue-600 underline">Dashboard</Link>
        <Link to="/auction" className="text-blue-600 underline">Auction</Link>
      </nav>
      <h1 className="text-3xl font-bold mb-6">Auction Demo</h1>
      <form onSubmit={handleSubmit} className="bg-white p-6 rounded shadow space-y-4">
        <input
          type="text"
          placeholder="Device ID"
          value={deviceId}
          onChange={(e) => setDeviceId(e.target.value)}
          className="w-full border p-2 rounded"
          required
        />
        <input
          type="text"
          placeholder="IP Address"
          value={ip}
          onChange={(e) => setIp(e.target.value)}
          className="w-full border p-2 rounded"
          required
        />
        <div className="flex gap-4">
          <input
            type="number"
            step="any"
            placeholder="Latitude"
            value={lat}
            onChange={(e) => setLat(e.target.value)}
            className="w-1/2 border p-2 rounded"
            required
          />
          <input
            type="number"
            step="any"
            placeholder="Longitude"
            value={lng}
            onChange={(e) => setLng(e.target.value)}
            className="w-1/2 border p-2 rounded"
            required
          />
        </div>
        <button type="submit" className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700">
          Send Bid
        </button>
      </form>
      {error && <p className="text-red-500 mt-4">Error: {error}</p>}
      {response && (
        <pre className="bg-gray-100 p-4 rounded mt-4 overflow-x-auto">
          {JSON.stringify(response, null, 2)}
        </pre>
      )}
    </div>
  );
}