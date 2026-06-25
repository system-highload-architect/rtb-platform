import { useEffect, useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';
import { fetchReport, fetchForecast, fetchFactorAnalysis, getExportUrl } from '../api/gateway';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const [report, setReport] = useState<any[]>([]);
  const [forecast, setForecast] = useState<number[]>([]);
  const [factor, setFactor] = useState<number[]>([]);
  const now = new Date();
  const start = '2026-06-01';
  const end = '2026-06-30';

  useEffect(() => {
    // Загружаем отчёт (пока пустой, но для демонстрации)
    fetchReport(start, end).then(setReport).catch(console.error);
    // Прогноз с тестовыми данными
    const history = [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15];
    fetchForecast(history, 3).then(data => setForecast(data.forecast)).catch(console.error);
    // Факторный анализ
    fetchFactorAnalysis().then(data => setFactor(data.explained_variance_ratio)).catch(console.error);
  }, []);

  return (
    <div className="p-6 max-w-6xl mx-auto">
        <nav className="mb-4">
            <Link to="/" className="mr-4 text-blue-600 underline">Dashboard</Link>
            <Link to="/auction" className="text-blue-600 underline">Auction</Link>
        </nav>
      <h1 className="text-3xl font-bold mb-6">RTB Platform Dashboard</h1>

      {/* Прогноз */}
      <div className="bg-white p-4 rounded shadow mb-6">
        <h2 className="text-xl font-semibold mb-4">Forecast</h2>
        {forecast.length > 0 ? (
          <LineChart width={600} height={300} data={forecast.map((v, i) => ({ step: i, value: v }))}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="step" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="value" stroke="#8884d8" />
          </LineChart>
        ) : (
          <p>Loading forecast...</p>
        )}
      </div>

      {/* Отчёт */}
      <div className="bg-white p-4 rounded shadow mb-6">
        <h2 className="text-xl font-semibold mb-4">Report</h2>
        {report.length > 0 ? (
          <table className="w-full border-collapse">
            <thead>
              <tr>
                {Object.keys(report[0].dimension_values || {}).map(dim => <th key={dim} className="border p-2">{dim}</th>)}
                {Object.keys(report[0].metric_values || {}).map(met => <th key={met} className="border p-2">{met}</th>)}
              </tr>
            </thead>
            <tbody>
              {report.map((row, i) => (
                <tr key={i}>
                  {Object.values(row.dimension_values || {}).map((val: any, j) => <td key={j} className="border p-2">{val}</td>)}
                  {Object.values(row.metric_values || {}).map((val: any, j) => <td key={j} className="border p-2">{val}</td>)}
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <p>No report data. Run an auction first.</p>
        )}
      </div>

      {/* Факторный анализ */}
      <div className="bg-white p-4 rounded shadow mb-6">
        <h2 className="text-xl font-semibold mb-4">Factor Analysis (Explained Variance)</h2>
        {factor.length > 0 ? (
          <ul>
            {factor.map((v, i) => <li key={i}>Component {i+1}: {v.toFixed(4)}</li>)}
          </ul>
        ) : (
          <p>No factor data</p>
        )}
      </div>

      {/* Экспорт */}
      <div className="bg-white p-4 rounded shadow">
        <h2 className="text-xl font-semibold mb-4">Export</h2>
        <a
          href={getExportUrl(start, end)}
          className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
          download
        >
          Download Excel Report
        </a>
      </div>
    </div>
  );
}