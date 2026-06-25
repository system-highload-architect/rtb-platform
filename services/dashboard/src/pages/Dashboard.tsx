import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';
import { fetchReport, fetchForecast, fetchFactorAnalysis, getExportUrl } from '../api/gateway';

export default function Dashboard() {
  const [report, setReport] = useState<any[]>([]);
  const [forecast, setForecast] = useState<number[]>([]);
  const [factor, setFactor] = useState<number[]>([]);

  // Состояния загрузки
  const [loadingReport, setLoadingReport] = useState(true);
  const [loadingForecast, setLoadingForecast] = useState(true);
  const [loadingFactor, setLoadingFactor] = useState(true);

  const now = new Date();
  const start = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`;
  const end = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-30`;

  useEffect(() => {
    // Загружаем отчёт
    fetchReport(start, end)
      .then(setReport)
      .catch(console.error)
      .finally(() => setLoadingReport(false));

    // Прогноз с тестовыми данными
    const history = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15];
    fetchForecast(history, 3)
      .then(data => setForecast(data.forecast))
      .catch(console.error)
      .finally(() => setLoadingForecast(false));

    // Факторный анализ
    fetchFactorAnalysis()
      .then(data => setFactor(data.explained_variance_ratio))
      .catch(console.error)
      .finally(() => setLoadingFactor(false));
  }, []);

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <nav className="mb-4">
        <Link to="/" className="mr-4 text-blue-600 underline">Dashboard</Link>
        <Link to="/auction" className="text-blue-600 underline">Auction</Link>
      </nav>
      <h1 className="text-3xl font-bold mb-6">RTB Platform Dashboard</h1>

      {/* Отчёт */}
      <div className="bg-white p-4 rounded shadow mb-6">
        <h2 className="text-xl font-semibold mb-4">Report</h2>
        {loadingReport ? (
          <div className="flex justify-center py-4">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : report.length > 0 ? (
          <table className="min-w-full border border-gray-200 divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                {Object.keys(report[0].dimension_values || {}).map(dim => (
                  <th key={dim} className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{dim}</th>
                ))}
                {Object.keys(report[0].metric_values || {}).map(met => (
                  <th key={met} className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{met}</th>
                ))}
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {report.map((row, i) => (
                <tr key={i} className={i % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                  {Object.values(row.dimension_values || {}).map((val: any, j) => (
                    <td key={j} className="px-4 py-3 whitespace-nowrap text-sm text-gray-700">{val}</td>
                  ))}
                  {Object.values(row.metric_values || {}).map((val: any, j) => (
                    <td key={j} className="px-4 py-3 whitespace-nowrap text-sm text-gray-700">{val}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <p>No report data. Run an auction first.</p>
        )}
      </div>

      {/* Прогноз */}
      <div className="bg-white p-4 rounded shadow mb-6">
        <h2 className="text-xl font-semibold mb-4">Forecast</h2>
        {loadingForecast ? (
          <div className="flex justify-center py-4">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : forecast.length > 0 ? (
          <LineChart width={600} height={300} data={forecast.map((v, i) => ({ step: i, value: v }))}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="step" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="value" stroke="#8884d8" />
          </LineChart>
        ) : (
          <p>No forecast data</p>
        )}
      </div>

      {/* Факторный анализ */}
      <div className="bg-white p-4 rounded shadow mb-6">
        <h2 className="text-xl font-semibold mb-4">Factor Analysis (Explained Variance)</h2>
        {loadingFactor ? (
          <div className="flex justify-center py-4">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          </div>
        ) : factor.length > 0 ? (
          <ul>
            {factor.map((v, i) => <li key={i}>Component {i + 1}: {v.toFixed(4)}</li>)}
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