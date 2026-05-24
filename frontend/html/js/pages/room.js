// frontend/js/pages/room.js
import { api } from '../api.js';
import { requireAuth } from '../auth.js';
import { renderHeader } from '../components.js';

// Require auth immediately
requireAuth();

const params = new URLSearchParams(location.search);
const roomId = parseInt(params.get('id') || '0');
if (!roomId) { alert('无效的房间 ID'); history.back(); }

let myChart = null;
let records = [];
let currentStartTime = null;
let currentEndTime = null;

// ——— ECharts 初始化 ———
function initChart() {
  myChart = echarts.init(document.getElementById('chart'), 'dark');
  myChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1a1d27',
      borderColor: '#2d3148',
      textStyle: { color: '#e8eaf6' },
      axisPointer: {
        type: 'cross',
        crossStyle: { color: '#8890b0' },
        lineStyle: { color: '#2d3148' }
      },
      formatter: function(params) {
        let html = `<div>${params[0].name}</div>`;
        let u = 0;
        params.forEach(p => {
          html += `<div style="margin-top:4px">${p.marker} ${p.seriesName}: <span style="font-weight:600">${p.value !== undefined ? p.value : 0}</span></div>`;
          if (p.seriesName === '用电量(度)') u = p.value || 0;
        });
        let cost = (u * 0.53).toFixed(2);
        html += `<div style="margin-top:4px"><span style="display:inline-block;margin-right:4px;border-radius:10px;width:10px;height:10px;background-color:#fb923c;"></span> 电费(元): <span style="font-weight:600">${cost}</span></div>`;
        return html;
      }
    },
    legend: { data: ['用电量(度)', '余额(元)'], textStyle: { color: '#8890b0' }, top: 8 },
    grid: { top: 50, bottom: 50, left: 50, right: 55, containLabel: false },
    xAxis: {
      type: 'category',
      data: [],
      boundaryGap: false,
      axisLabel: { color: '#8890b0', fontSize: 11, rotate: 0 },
      axisLine: { lineStyle: { color: '#2d3148' } },
      axisTick: { alignWithLabel: true }
    },
    yAxis: [
      {
        type: 'value', name: '度',
        nameTextStyle: { color: '#8890b0' },
        splitLine: { lineStyle: { color: '#2d3148' } },
        axisLabel: { color: '#8890b0' },
        min: 0
      },
      {
        type: 'value', name: '元',
        nameTextStyle: { color: '#8890b0' },
        splitLine: { show: false },
        axisLabel: { color: '#8890b0' }
      },
    ],
    series: [
      {
        name: '用电量(度)', type: 'line', data: [], smooth: true,
        itemStyle: { color: '#4f9cf9' },
        lineStyle: { width: 2, color: '#4f9cf9' },
        symbol: 'circle', symbolSize: 6, showSymbol: false,
        emphasis: { focus: 'series', itemStyle: { color: '#4f9cf9', borderColor: '#fff', borderWidth: 2 } },
        areaStyle: {
          color: { type: 'linear', x:0,y:0,x2:0,y2:1, colorStops:[{offset:0,color:'rgba(79,156,249,0.2)'},{offset:1,color:'transparent'}] }
        }
      },
      {
        name: '余额(元)', type: 'line', yAxisIndex: 1, data: [], smooth: true,
        itemStyle: { color: '#34d399' },
        lineStyle: { color: '#34d399', width: 2 },
        symbol: 'circle', symbolSize: 6, showSymbol: false,
        emphasis: { focus: 'series', itemStyle: { color: '#34d399', borderColor: '#fff', borderWidth: 2 } },
        areaStyle: {
          color: { type: 'linear', x:0,y:0,x2:0,y2:1, colorStops:[{offset:0,color:'rgba(52,211,153,0.2)'},{offset:1,color:'transparent'}] }
        }
      },
    ],
  });
  window.addEventListener('resize', () => myChart?.resize());
}

// ——— 数据聚合 ———
function round2(n) { return Math.round(n * 100) / 100; }

function calcTimeParams(spanHours) {
  const MAX_TICKS = 35;
  let bucketHours, labelFn;

  if (spanHours <= 24) {
    bucketHours = 1;
    labelFn = (d) => `${String(d.getHours()).padStart(2,'0')}:00`;
  } else if (spanHours <= 24 * 7) {
    const ticksAt1h = Math.ceil(spanHours / 1);
    if (ticksAt1h <= MAX_TICKS) {
      bucketHours = 1;
    } else {
      bucketHours = Math.ceil(spanHours / MAX_TICKS);
      const niceValues = [1, 2, 3, 4, 6, 8, 12];
      for (const nv of niceValues) {
        if (nv >= bucketHours) { bucketHours = nv; break; }
      }
    }
    labelFn = (d) => `${d.getMonth()+1}/${d.getDate()} ${String(d.getHours()).padStart(2,'0')}:00`;
  } else {
    const spanDays = spanHours / 24;
    const ticksAt1d = Math.ceil(spanDays);
    if (ticksAt1d <= MAX_TICKS) {
      bucketHours = 24;
    } else {
      const bucketDays = Math.ceil(spanDays / MAX_TICKS);
      bucketHours = bucketDays * 24;
    }
    labelFn = (d) => `${d.getMonth()+1}/${d.getDate()}`;
  }
  return { bucketHours, labelFn };
}

function bucketStart(ms, bucketHours) {
  const d = new Date(ms);
  if (bucketHours >= 24) {
    d.setHours(0,0,0,0);
  } else {
    const h = Math.floor(d.getHours() / bucketHours) * bucketHours;
    d.setHours(h, 0, 0, 0);
  }
  return d.getTime();
}

function buildChart(rows) {
  if (!rows.length) return { labels:[], usage:[], balance:[], balanceMin:undefined, balanceMax:undefined };

  const firstTime = new Date(rows[0].end_time).getTime();
  const lastTime = new Date(rows[rows.length-1].end_time).getTime();
  const spanH = (lastTime - firstTime) / 3600000;

  const { bucketHours, labelFn } = calcTimeParams(spanH || 1);

  const map = new Map();
  for (const row of rows) {
    const t = new Date(row.end_time).getTime();
    const bkt = bucketStart(t, bucketHours);
    if (!map.has(bkt)) map.set(bkt, { usage: 0, balance: 0, t: bkt });
    map.get(bkt).usage += row.usage;
    map.get(bkt).balance = row.balance;
  }
  const pts = Array.from(map.values()).sort((a,b)=>a.t-b.t);

  const labels = pts.map(p => labelFn(new Date(p.t)));
  const usage = pts.map(p => round2(p.usage));
  const balance = pts.map(p => round2(p.balance));
  const bVals = balance.filter(v => v > 0);
  let balanceMin, balanceMax;
  if (bVals.length) {
    const bMin = Math.min(...bVals), bMax = Math.max(...bVals);
    const pad = Math.max((bMax-bMin)*0.3, 2);
    balanceMin = Math.floor(bMin - pad);
    balanceMax = Math.ceil(bMax + pad);
  }
  return { labels, usage, balance, balanceMin, balanceMax };
}

function updateChart() {
  const overlay = document.getElementById('chartOverlay');
  if (!records.length) {
    overlay.style.display = 'flex';
    myChart?.setOption({ xAxis: { data: [] }, series: [{ data: [] }, { data: [] }] });
    return;
  }
  overlay.style.display = 'none';
  const { labels, usage, balance, balanceMin, balanceMax } = buildChart(records);

  let labelInterval = 0;
  if (labels.length > 35) {
    labelInterval = Math.ceil(labels.length / 35) - 1;
  }

  myChart.setOption({
    xAxis: {
      data: labels,
      axisLabel: {
        interval: labelInterval,
        rotate: labels.length > 20 ? 20 : 0
      }
    },
    yAxis: [ {}, { min: balanceMin, max: balanceMax } ],
    series: [{ data: usage }, { data: balance }],
  });

  const total = usage.reduce((s,v)=>s+v, 0);
  document.getElementById('statUsage').textContent = `${round2(total).toFixed(2)} 度`;
  document.getElementById('statCost').textContent = `${(total * 0.53).toFixed(2)} 元`;
}

// ——— 数据加载 ———
function formatRangeDate(isoStr) {
  if (!isoStr) return '—';
  const d = new Date(isoStr);
  return `${d.getFullYear()}/${String(d.getMonth()+1).padStart(2,'0')}/${String(d.getDate()).padStart(2,'0')} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`;
}

async function loadUsage(start, end) {
  currentStartTime = start;
  currentEndTime = end;
  
  const elRange = document.getElementById('statTimeRange');
  elRange.innerHTML = `${formatRangeDate(start)}<br>${formatRangeDate(end)}`;

  const errEl = document.getElementById('errorMsg');
  errEl.style.display = 'none';
  try {
    const res = await api.get(`/electricity-balances/hour-range/${roomId}`, {
      params: { start_time: start, end_time: end }
    });
    records = res.data || [];
    updateChart();
  } catch (e) {
    errEl.textContent = '加载用电数据失败';
    errEl.style.display = 'block';
    records = [];
    updateChart();
  }
}

// ——— 快捷时间 ———
const RANGES = {
  today:      () => { const s = startOfDay(new Date()); return [s, new Date()]; },
  yesterday:  () => { const d = new Date(); d.setDate(d.getDate()-1); return [startOfDay(d), endOfDay(d)]; },
  this_week:  () => { const d = new Date(); d.setDate(d.getDate()-d.getDay()); return [startOfDay(d), new Date()]; },
  last_week:  () => { const d = new Date(); d.setDate(d.getDate()-d.getDay()-7); const e = new Date(d); e.setDate(d.getDate()+6); return [startOfDay(d), endOfDay(e)]; },
  this_month: () => { const d = new Date(); d.setDate(1); return [startOfDay(d), new Date()]; },
  last_month: () => { const d = new Date(); d.setDate(1); d.setMonth(d.getMonth()-1); const e = new Date(d.getFullYear(), d.getMonth()+1, 0); return [startOfDay(d), endOfDay(e)]; },
};
function startOfDay(d) { const r=new Date(d); r.setHours(0,0,0,0); return r; }
function endOfDay(d)   { const r=new Date(d); r.setHours(23,59,59,999); return r; }

document.getElementById('quickBtns').addEventListener('click', e => {
  const btn = e.target.closest('.qbtn');
  if (!btn) return;
  document.querySelectorAll('.qbtn').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  const [s, end] = RANGES[btn.dataset.v]();
  loadUsage(s.toISOString(), end.toISOString());
});

let dateRangePicker = null;

window.doCustomSearch = function() {
  if (!dateRangePicker) return;
  const dates = dateRangePicker.selectedDates;
  if (dates.length !== 2) { alert('请选择开始和结束日期'); return; }
  document.querySelectorAll('.qbtn').forEach(b => b.classList.remove('active'));
  loadUsage(startOfDay(dates[0]).toISOString(), endOfDay(dates[1]).toISOString());
}

window.doClearDateRange = function() {
  if (!dateRangePicker) return;
  dateRangePicker.clear();
  document.querySelectorAll('.qbtn').forEach(b => b.classList.remove('active'));
  document.querySelector('.qbtn[data-v="today"]').classList.add('active');
  const [s, end] = RANGES['today']();
  loadUsage(s.toISOString(), end.toISOString());
}

// ——— 初始化 ———
async function init() {
  renderHeader();
  initChart();
  
  if (window.flatpickr) {
    dateRangePicker = window.flatpickr("#dateRange", {
      mode: "range",
      dateFormat: "Y-m-d",
      locale: "zh",
      maxDate: "today"
    });
  }
  
  try {
    const [ri, bi] = await Promise.all([
      api.get(`/rooms/${roomId}`),
      api.get(`/electricity-balances/latest/${roomId}`).catch(() => null),
    ]);
    const room = ri.data;
    document.title = `${room.name} — ElecLog`;
    document.getElementById('headerTitle').textContent = room.name;
    document.getElementById('statRoom').textContent = room.name;

    if (bi) {
      const yuan = bi.data.balance / 100;
      const balEl = document.getElementById('statBalance');
      balEl.textContent = `¥ ${yuan.toFixed(2)}`;
      balEl.className = 'stat-value ' + (yuan >= 15 ? 'good' : yuan >= 8 ? 'warn' : 'low');
    }
  } catch (e) {
    document.getElementById('errorMsg').textContent = '加载房间信息失败';
    document.getElementById('errorMsg').style.display = 'block';
  }
  const [s, end] = RANGES['today']();
  loadUsage(s.toISOString(), end.toISOString());
}

document.addEventListener('DOMContentLoaded', init);
