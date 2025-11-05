const routes = ['dashboard', 'recommendations', 'impacted', 'resourceType', 'inventory', 'advisor', 'azurePolicy', 'arcSQL', 'defender', 'defenderRecommendations', 'costs', 'outOfScope'];
const routeLabels = {
    'dashboard': 'Overview',
    'recommendations': 'Recommendations',
    'impacted': 'Impacted Resources',
    'resourceType': 'Resource Types',
    'inventory': 'Resource Inventory',
    'advisor': 'Azure Advisor',
    'azurePolicy': 'Azure Policy',
    'arcSQL': 'Arc SQL Server',
    'defender': 'Microsoft Defender',
    'defenderRecommendations': 'Defender Recommendations',
    'costs': 'Cost Analysis',
    'outOfScope': 'Out of Scope'
};
let pluginRoutes = [];
let pluginMetadata = {};
const menuStructure = [
    { label: 'Overview', route: 'dashboard', children: [] },
    {
        label: 'Best Practices',
        children: [
            { label: 'Recommendations', route: 'recommendations' },
            { label: 'Impacted Resources', route: 'impacted' },
            { label: 'Azure Advisor', route: 'advisor' }
        ]
    },
    {
        label: 'Inventory',
        children: [
            { label: 'Resource Types', route: 'resourceType' },
            { label: 'Resource Inventory', route: 'inventory' },
            { label: 'Out of Scope', route: 'outOfScope' }
        ]
    },
    {
        label: 'Security',
        children: [
            { label: 'Microsoft Defender', route: 'defender' },
            { label: 'Defender Recommendations', route: 'defenderRecommendations' }
        ]
    },
    {
        label: 'Governance',
        children: [
            { label: 'Azure Policy', route: 'azurePolicy' },
            { label: 'Cost Analysis', route: 'costs' }
        ]
    },
    {
        label: 'Arc',
        children: [
            { label: 'SQL Server', route: 'arcSQL' }
        ]
    }
];
function showLoading(message = 'Loading data...') {
    const overlay = document.getElementById('loading-overlay');
    const text = overlay.querySelector('.loading-text');
    if (text) text.textContent = message;
    overlay.style.display = 'flex';
}

function hideLoading() {
    const overlay = document.getElementById('loading-overlay');
    overlay.style.display = 'none';
}

async function fetchJSON(u) { const r = await fetch(u); if (!r.ok) throw new Error(await r.text()); return r.json(); }
function navigate() {
    const h = location.hash.replace('#', '');
    console.log('navigate called with hash:', h);
    if (!h || h === 'dashboard') {
        updateBrand('dashboard');
        showDashboard();
        return;
    }
    if (routes.includes(h)) {
        updateBrand(h);
        showDataset(h);
    } else if (pluginRoutes.includes(h)) {
        updateBrand(h);
        showPlugin(h);
    } else {
        console.log('Unknown route:', h, 'redirecting to dashboard');
        updateBrand('dashboard');
        showDashboard();
    }
}

function updateBrand(route) {
    const brand = document.getElementById('brand-text');
    // Always show the route label
    const label = routeLabels[route] || 'Overview';
    const title = `Azure Quick Review – ${label}`;

    // Update navbar brand text
    if (brand) {
        brand.textContent = title;
    }

    // Update browser tab title
    document.title = `azqr - ${label}`;
}
async function initNav() {
    const nav = document.getElementById('nav');
    let html = '';

    menuStructure.forEach(section => {
        if (section.route) {
            // Single item (like Home)
            html += `<li class="nav-item">
				<a class="nav-link" href="#${section.route}" onclick="closeMenu()">
					<i class="bi bi-house-door me-1"></i>${section.label}
				</a>
			</li>`;
        } else if (section.children && section.children.length > 0) {
            // Group with children - use Bootstrap dropdown
            html += `<li class="nav-item dropdown">
				<a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown">
					${getMenuIcon(section.label)} ${section.label}
				</a>
				<ul class="dropdown-menu">`;
            section.children.forEach(child => {
                html += `<li><a class="dropdown-item" href="#${child.route}" onclick="closeMenu()">
					${getMenuIcon(child.label)} ${child.label}
				</a></li>`;
            });
            html += `</ul></li>`;
        }
    });

    // Load plugins and add plugin menu if plugins exist
    try {
        const plugins = await fetchJSON('/api/plugins');
        if (plugins && plugins.length > 0) {
            pluginRoutes = plugins.map(p => `plugin-${p.name}`);
            plugins.forEach(p => {
                pluginMetadata[`plugin-${p.name}`] = p;
                routeLabels[`plugin-${p.name}`] = p.displayName || p.name;
            });

            html += `<li class="nav-item dropdown">
                <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown">
                    <i class="bi bi-plugin me-1"></i> Plugins
                </a>
                <ul class="dropdown-menu">`;
            plugins.forEach(p => {
                html += `<li><a class="dropdown-item" href="#plugin-${p.name}" onclick="closeMenu()">
                    <i class="bi bi-puzzle me-1"></i> ${p.displayName || p.name}
                </a></li>`;
            });
            html += `</ul></li>`;
        }
    } catch (error) {
        console.warn('No plugins available or error loading plugins:', error);
    }

    nav.innerHTML = html;
}

function getMenuIcon(label) {
    const icons = {
        'Overview': '<i class="bi bi-house-door me-1"></i>',
        'Best Practices': '<i class="bi bi-lightbulb me-1"></i>',
        'Recommendations': '<i class="bi bi-lightbulb me-1"></i>',
        'Impacted Resources': '<i class="bi bi-exclamation-triangle me-1"></i>',
        'Azure Advisor': '<i class="bi bi-person-check me-1"></i>',
        'Inventory': '<i class="bi bi-archive me-1"></i>',
        'Resource Types': '<i class="bi bi-collection me-1"></i>',
        'Resource Inventory': '<i class="bi bi-box-seam me-1"></i>',
        'Out of Scope': '<i class="bi bi-x-circle me-1"></i>',
        'Arc': '<i class="bi bi-hdd-network me-1"></i>',
        'SQL Server': '<i class="bi bi-database me-1"></i>',
        'Security': '<i class="bi bi-shield-lock me-1"></i>',
        'Microsoft Defender': '<i class="bi bi-shield-check me-1"></i>',
        'Defender Recommendations': '<i class="bi bi-shield-exclamation me-1"></i>',
        'Governance': '<i class="bi bi-bank me-1"></i>',
        'Azure Policy': '<i class="bi bi-file-earmark-ruled me-1"></i>',
        'Cost Analysis': '<i class="bi bi-graph-up me-1"></i>'
    };
    return icons[label] || '<i class="bi bi-circle me-1"></i>';
} function toggleMenu() {
    // Bootstrap handles the navbar collapse automatically
}

function closeMenu() {
    // Close Bootstrap navbar if it's open
    const navbarCollapse = document.getElementById('navbarNav');
    const bsCollapse = bootstrap.Collapse.getInstance(navbarCollapse);
    if (bsCollapse) {
        bsCollapse.hide();
    }
}
async function showDashboard() {
    document.getElementById('dashboard').style.display = 'block';
    document.getElementById('dataset-view').style.display = 'none';
    const cards = document.getElementById('cards');

    showLoading('Loading dashboard...');
    let summary, analytics;
    try {
        [summary, analytics] = await Promise.all([
            fetchJSON('/api/summary'),
            fetchJSON('/api/analytics').catch(() => ({}))
        ]);
    } catch (error) {
        hideLoading();
        console.error('Error loading dashboard:', error);
        cards.innerHTML = '<div class="col-12"><div class="alert alert-danger"><i class="bi bi-exclamation-triangle me-2"></i>Error loading dashboard data</div></div>';
        return;
    }

    cards.innerHTML = '';
    // Basic summary cards (filter out low-value counts per request)
    const exclude = new Set([
        'advisorCount', 'azurePolicyCount', 'costItems', 'defenderCount', 'defenderRecommendationsCount',
        'impactedCount', 'inventoryCount', 'outOfScopeCount', 'azurePolicyNonCompliant',
        'recommendationsImplemented', 'recommendationsNotImplemented', 'recommendationsTotal',
        'resourceTypeCount', 'totalCost', 'arcSQLCount', 'pluginCounts'
    ]);

    Object.entries(summary).forEach(([k, v]) => {
        if (exclude.has(k)) return;
        const target = routes.find(r => k.toLowerCase().includes(r.toLowerCase()));
        const link = target ? `#${target}` : '#';
        const cardIcon = getCardIcon(k);
        const cardColor = getCardColor(k);
        cards.innerHTML += `
			<div class="col-lg-3 col-md-4 col-sm-6">
				<div class="card h-100 border-0 shadow-sm">
					<div class="card-body d-flex align-items-center">
						<div class="flex-shrink-0 me-3">
							<div class="rounded-circle bg-${cardColor} p-3 text-white">
								${cardIcon}
							</div>
						</div>
						<div class="flex-grow-1">
							<h6 class="card-title text-muted mb-1 small">${makeReadable(k)}</h6>
							<h3 class="card-text fw-bold mb-0 text-${cardColor}">${v}</h3>
							${target ? `<a href="${link}" class="btn btn-outline-${cardColor} btn-sm mt-2">View Details</a>` : ''}
						</div>
					</div>
				</div>
			</div>`;
    });

    // Top tables (categories & resource types & impact distribution & defender severity) appended below cards
    const existingTables = document.getElementById('analytics-tables');
    if (existingTables) existingTables.remove();
    const wrapper = document.createElement('div');
    wrapper.id = 'analytics-tables';
    wrapper.className = 'row mt-4';

    // Impact Distribution as table with pie chart
    if (analytics.impact && analytics.impact.distribution) {
        const dist = analytics.impact.distribution;
        const distRows = [
            { Impact: 'High', Count: dist.high || 0 },
            { Impact: 'Medium', Count: dist.medium || 0 },
            { Impact: 'Low', Count: dist.low || 0 }
        ];
        const total = (dist.high || 0) + (dist.medium || 0) + (dist.low || 0);
        wrapper.appendChild(buildBootstrapTableWithChart('Impact Distribution', distRows, ['Impact', 'Count'], total, 'danger'));
    }

    // Defender Severity as table with pie chart
    if (analytics.defender && analytics.defender.severity) {
        const sev = analytics.defender.severity;
        const sevRows = [
            { Severity: 'High', Count: sev.high || 0 },
            { Severity: 'Medium', Count: sev.medium || 0 },
            { Severity: 'Low', Count: sev.low || 0 }
        ];
        const total = (sev.high || 0) + (sev.medium || 0) + (sev.low || 0);
        wrapper.appendChild(buildBootstrapTableWithChart('Defender Severity', sevRows, ['Severity', 'Count'], total, 'warning'));
    }

    if (analytics.categories && analytics.categories.top) {
        const readableCats = analytics.categories.top.map(cat => ({
            'Category': cat.category,
            'Total': cat.impactedTotal,
            'High': cat.highImpact,
            'Medium': cat.mediumImpact,
            'Low': cat.lowImpact
        }));
        wrapper.appendChild(buildBootstrapTable('Top Categories (Impacted)', readableCats, ['Category', 'Total', 'High', 'Medium', 'Low'], 'info'));
    }
    if (analytics.resourceTypes && analytics.resourceTypes.topImpacted) {
        const readableRTs = analytics.resourceTypes.topImpacted.map(rt => ({
            'Resource Type': rt.resourceType,
            'Count': rt.impactedCount
        }));
        wrapper.appendChild(buildBootstrapTable('Top Impacted Resource Types', readableRTs, ['Resource Type', 'Count'], 'success'));
    }
    cards.parentElement.appendChild(wrapper);

    hideLoading();
}

function getCardIcon(key) {
    const icons = {
        'recommendationsTotal': '<i class="bi bi-lightbulb fs-4"></i>',
        'impactedCount': '<i class="bi bi-exclamation-triangle fs-4"></i>',
        'resourceTypeCount': '<i class="bi bi-collection fs-4"></i>',
        'inventoryCount': '<i class="bi bi-box-seam fs-4"></i>',
        'advisorCount': '<i class="bi bi-person-check fs-4"></i>',
        'azurePolicyCount': '<i class="bi bi-file-earmark-ruled fs-4"></i>',
        'defenderCount': '<i class="bi bi-shield-check fs-4"></i>',
        'costItems': '<i class="bi bi-graph-up fs-4"></i>'
    };
    return icons[key] || '<i class="bi bi-circle fs-4"></i>';
}

function getCardColor(key) {
    const colors = {
        'recommendationsTotal': 'primary',
        'impactedCount': 'danger',
        'resourceTypeCount': 'info',
        'inventoryCount': 'success',
        'advisorCount': 'warning',
        'azurePolicyCount': 'secondary',
        'defenderCount': 'dark',
        'costItems': 'primary'
    };
    return colors[key] || 'primary';
}

function card(title, value) { return `<div class=\"card\"><h3>${title}</h3><p>${escapeHTML(String(value))}</p></div>` }
function pct(v) { return (v || 0).toFixed(1) }
function buildBootstrapTable(title, rows, cols, color = 'primary') {
    const div = document.createElement('div');
    div.className = 'col-lg-6 col-md-12 mb-4';
    let html = `
		<div class="card h-100 border-0 shadow-sm">
			<div class="card-header bg-${color} text-white">
				<h6 class="card-title mb-0 fw-bold">${title}</h6>
			</div>
			<div class="card-body p-0">
				<div class="table-responsive">
					<table class="table table-sm mb-0">
						<thead class="table-light">
							<tr>${cols.map(c => `<th class="fw-semibold">${c}</th>`).join('')}</tr>
						</thead>
						<tbody>`;
    rows.forEach(r => {
        html += '<tr>' + cols.map(c => `<td>${escapeHTML(String(r[c] ?? ''))}</td>`).join('') + '</tr>';
    });
    html += `</tbody></table></div></div></div>`;
    div.innerHTML = html;
    return div;
}

function buildBootstrapTableWithChart(title, rows, cols, total, color = 'primary') {
    const div = document.createElement('div');
    div.className = 'col-lg-6 col-md-12 mb-4';
    let html = `
		<div class="card h-100 border-0 shadow-sm">
			<div class="card-header bg-${color} text-white">
				<h6 class="card-title mb-0 fw-bold">${title}</h6>
			</div>
			<div class="card-body">
				<div class="row align-items-center">
					<div class="col-8">
						<div class="table-responsive">
							<table class="table table-sm mb-0">
								<thead class="table-light">
									<tr>${cols.map(c => `<th class="fw-semibold">${c}</th>`).join('')}</tr>
								</thead>
								<tbody>`;
    rows.forEach(r => {
        html += '<tr>' + cols.map(c => `<td>${escapeHTML(String(r[c] ?? ''))}</td>`).join('') + '</tr>';
    });
    html += `</tbody></table></div></div><div class="col-4 text-center">`;

    if (total > 0 && rows.length === 3) {
        const colors = ['#dc3545', '#fd7e14', '#198754']; // Bootstrap colors: danger, warning, success
        const counts = rows.map(r => r[cols[1]] || 0);
        const percents = counts.map(c => (c / total) * 100);
        let gradientStops = [];
        let cumulative = 0;
        for (let i = 0; i < percents.length; i++) {
            if (percents[i] > 0) {
                gradientStops.push(`${colors[i]} ${cumulative}% ${cumulative + percents[i]}%`);
                cumulative += percents[i];
            }
        }
        const gradientStyle = `conic-gradient(${gradientStops.join(', ')})`;
        html += `<div class="mx-auto rounded-circle border" style="width: 80px; height: 80px; background: ${gradientStyle};" title="High:${counts[0]} Medium:${counts[1]} Low:${counts[2]}"></div>`;
    }
    html += `</div></div></div></div>`;
    div.innerHTML = html;
    return div;
}
function buildTableWithPie(title, rows, cols, total) {
    const div = document.createElement('div');
    div.className = 'mini-table';
    let html = `<h3>${title}</h3><div class="table-pie-container"><table><thead><tr>${cols.map(c => `<th>${c}</th>`).join('')}</tr></thead><tbody>`;
    rows.forEach(r => { html += '<tr>' + cols.map(c => `<td>${escapeHTML(String(r[c] ?? ''))}</td>`).join('') + '</tr>'; });
    html += '</tbody></table>';
    if (total > 0 && rows.length === 3) {
        const colors = ['#e74c3c', '#f39c12', '#27ae60'];
        const counts = rows.map(r => r[cols[1]] || 0);
        const percents = counts.map(c => (c / total) * 100);
        let gradientStops = [];
        let cumulative = 0;
        for (let i = 0; i < percents.length; i++) {
            if (percents[i] > 0) {
                gradientStops.push(`${colors[i]} ${cumulative}% ${cumulative + percents[i]}%`);
                cumulative += percents[i];
            }
        }
        const gradientStyle = `conic-gradient(${gradientStops.join(', ')})`;
        html += `<div class="pie-chart" style="background: ${gradientStyle};" title="High:${counts[0]} Medium:${counts[1]} Low:${counts[2]}"></div>`;
    }
    html += '</div>';
    div.innerHTML = html;
    return div;
}
async function showDataset(name) {
    console.log('showDataset called with:', name);
    document.getElementById('dashboard').style.display = 'none';
    document.getElementById('dataset-view').style.display = 'block';

    showLoading(`Loading ${routeLabels[name] || name}...`);
    try {
        const data = await fetchJSON(`/api/data/${name}`);
        console.log('Data fetched:', data.length, 'rows');
        renderTable(data, name);
    } catch (error) {
        console.error('Error fetching data:', error);
        document.getElementById('data-table').innerHTML = '<tbody><tr><td class="text-center text-danger py-4"><i class="bi bi-exclamation-triangle me-2"></i>Error loading data</td></tr></tbody>';
    } finally {
        hideLoading();
    }
}

async function showPlugin(routeName) {
    console.log('showPlugin called with:', routeName);
    document.getElementById('dashboard').style.display = 'none';
    document.getElementById('dataset-view').style.display = 'block';

    const pluginName = routeName.replace('plugin-', '');
    const plugin = pluginMetadata[routeName];

    showLoading(`Loading ${plugin?.displayName || pluginName}...`);
    try {
        const data = await fetchJSON(`/api/plugin/${pluginName}`);
        console.log('Plugin data fetched:', data.length, 'rows');
        renderPluginTable(data, plugin);
    } catch (error) {
        console.error('Error fetching plugin data:', error);
        document.getElementById('data-table').innerHTML = '<tbody><tr><td class="text-center text-danger py-4"><i class="bi bi-exclamation-triangle me-2"></i>Error loading plugin data</td></tr></tbody>';
    } finally {
        hideLoading();
    }
}
function renderTable(rows, datasetName) {
    const table = document.getElementById('data-table');
    if (rows.length === 0) {
        table.innerHTML = '<tbody><tr><td class="text-center text-muted py-4"><i class="bi bi-inbox me-2"></i>No data available</td></tr></tbody>';
        return;
    }

    // Define column orders matching report_data.go table functions
    const columnOrders = {
        'impacted': ['validatedUsing', 'source', 'category', 'impact', 'resourceType', 'recommendation', 'recommendationId', 'subscriptionId', 'subscriptionName', 'resourceGroup', 'resourceName', 'resourceId', 'param1', 'param2', 'param3', 'param4', 'param5', 'learn'],
        'costs': ['from', 'to', 'subscriptionId', 'subscriptionName', 'serviceName', 'value', 'currency'],
        'defender': ['subscriptionId', 'subscriptionName', 'name', 'tier'],
        'advisor': ['subscriptionId', 'subscriptionName', 'resourceType', 'resourceName', 'category', 'impact', 'description', 'resourceId', 'recommendationId'],
        'azurePolicy': ['subscriptionId', 'subscriptionName', 'resourceGroup', 'resourceType', 'resourceName', 'policyDisplayName', 'policyDescription', 'resourceId', 'timeStamp', 'policyDefinitionName', 'policyDefinitionId', 'policyAssignmentName', 'policyAssignmentId', 'complianceState'],
        'arcSQL': ['subscriptionId', 'subscriptionName', 'azureArcServer', 'sqlInstance', 'resourceGroup', 'version', 'build', 'patchLevel', 'edition', 'vCores', 'dpsStatus', 'license', 'telStatus', 'defenderStatus'],
        'recommendations': ['implemented', 'numberOfImpactedResources', 'azureServiceWellArchitected', 'recommendationSource', 'azureServiceCategoryWellArchitectedArea', 'azureServiceWellArchitectedTopic', 'category', 'recommendation', 'impact', 'bestPracticesGuidance', 'readMore', 'recommendationId'],
        'resourceType': ['subscription', 'resourceType', 'numberOfResources'],
        'defenderRecommendations': ['subscriptionId', 'subscriptionName', 'resourceGroup', 'resourceType', 'resourceName', 'category', 'recommendationSeverity', 'recommendationName', 'actionDescription', 'remediationDescription', 'resourceId'],
        'inventory': ['subscriptionId', 'resourceGroup', 'location', 'resourceType', 'resourceName', 'skuName', 'skuTier', 'kind', 'sla', 'resourceId'],
        'outOfScope': ['subscriptionId', 'resourceGroup', 'location', 'resourceType', 'resourceName', 'resourceId']
    };

    const allHeaders = Object.keys(rows[0]);
    const orderedHeaders = columnOrders[datasetName] || allHeaders;
    const finalHeaders = orderedHeaders.filter(h => allHeaders.includes(h));
    const readableHeaders = finalHeaders.map(h => makeReadable(h));

    // Define URL fields that should be rendered as clickable links
    const urlFields = ['learn', 'readMore', 'azPortalLink', 'learnMoreUrl'];

    // Build table with filter headers
    let html = '<thead class="table-dark sticky-top">';
    html += '<tr>' + readableHeaders.map(h => `<th scope="col" class="fw-semibold">${h}</th>`).join('') + '</tr>';

    // Add filter row
    html += '<tr class="table-secondary">';
    html += finalHeaders.map((field, index) => {
        // Define filterable columns per dataset
        const filterConfig = {
            'costs': ['subscriptionId', 'subscriptionName', 'serviceName'],
            'azurePolicy': ['subscriptionId', 'subscriptionName', 'resourceType', 'resourceName'],
            'arcSQL': ['subscriptionId', 'subscriptionName', 'resourceGroup', 'version', 'edition', 'dpsStatus', 'license', 'telStatus', 'defenderStatus'],
            'defender': ['subscriptionId', 'subscriptionName', 'name', 'tier'],
            'defenderRecommendations': ['subscriptionId', 'subscriptionName', 'resourceGroup', 'resourceType', 'resourceName', 'category', 'recommendationSeverity'],
            'inventory': ['subscriptionId', 'resourceGroup', 'location', 'resourceType', 'resourceName'],
            'outOfScope': ['subscriptionId', 'resourceGroup', 'location', 'resourceType', 'resourceName'],
            'resourceType': ['subscription', 'resourceType'],
            'advisor': ['subscriptionId', 'subscriptionName', 'resourceType', 'resourceName', 'category', 'impact', 'recommendationId'],
            'impacted': ['source', 'category', 'impact', 'resourceType', 'recommendationId', 'subscriptionId', 'subscriptionName', 'resourceGroup', 'resourceName'],
            'recommendations': ['recommendationId', 'subscriptionId', 'subscriptionName', 'resourceGroup', 'resourceName', 'implemented', 'recommendationSource', 'azureServiceCategoryWellArchitectedArea', 'azureServiceWellArchitectedTopic', 'category', 'impact']
        };

        // Define dropdown fields per dataset
        const dropdownConfig = {
            'advisor': ['resourceType', 'category', 'impact'],
            'impacted': ['source', 'resourceType', 'category', 'impact'],
            'defender': ['tier'],
            'recommendations': ['implemented', 'recommendationSource', 'azureServiceCategoryWellArchitectedArea', 'azureServiceWellArchitectedTopic', 'impact', 'category'],
            'defenderRecommendations': ['resourceType', 'category', 'recommendationSeverity'],
            'inventory': ['resourceType', 'location'],
            'outOfScope': ['resourceType', 'location'],
            'azurePolicy': ['resourceType'],
            'arcSQL': ['edition', 'dpsStatus', 'license', 'defenderStatus'],
        };

        const datasetFilters = filterConfig[datasetName] || [];
        const datasetDropdowns = dropdownConfig[datasetName] || [];

        if (datasetFilters.includes(field)) {
            // Get unique values using case-insensitive comparison
            const valueMap = new Map();
            rows.forEach(r => {
                const val = r[field];
                if (val && val.trim()) {
                    const lowerKey = val.toLowerCase();
                    // Keep first occurrence (preserves original casing)
                    if (!valueMap.has(lowerKey)) {
                        valueMap.set(lowerKey, val);
                    }
                }
            });
            const uniqueValues = Array.from(valueMap.values()).sort((a, b) =>
                a.toLowerCase().localeCompare(b.toLowerCase())
            );

            if (datasetDropdowns.includes(field) && uniqueValues.length >= 1 && uniqueValues.length <= 50) {
                // Dropdown for specified fields (show even with 1 value)
                return `<th class="p-1">
                    <select class="form-select form-select-sm table-filter" data-column="${index}">
                        <option value="">All</option>
                        ${uniqueValues.map(v => `<option value="${escapeHTML(v)}">${escapeHTML(v)}</option>`).join('')}
                    </select>
                </th>`;
            } else {
                // Text input for other filterable fields
                return `<th class="p-1">
                    <input type="text" class="form-control form-control-sm table-filter" data-column="${index}" placeholder="Filter...">
                </th>`;
            }
        }

        // No filter for non-filterable fields
        return `<th class="p-1"></th>`;
    }).join('');
    html += '</tr></thead>';

    // Build table body with filterable rows
    html += '<tbody>';
    rows.forEach((r, rowIndex) => {
        html += `<tr class="filterable-row ${rowIndex % 2 === 0 ? 'table-light' : ''}">`;
        html += finalHeaders.map(h => {
            const value = r[h] || '';
            if (urlFields.includes(h) && value && isValidUrl(value)) {
                return `<td><a href="${escapeHTML(value)}" target="_blank" rel="noopener noreferrer" class="btn btn-sm btn-outline-primary"><i class="bi bi-box-arrow-up-right me-1"></i>Learn More</a></td>`;
            }
            if (h === 'impact') {
                const badgeClass = value.toLowerCase() === 'high' ? 'danger' : value.toLowerCase() === 'medium' ? 'warning' : 'success';
                return `<td><span class="badge bg-${badgeClass}">${escapeHTML(value)}</span></td>`;
            }
            // Implemented status formatting only for recommendations dataset
            if (datasetName === 'recommendations' && h === 'implemented') {
                const raw = value.toString().trim().toLowerCase();
                let badgeClass = 'secondary';
                let label = value || 'N/A';
                if (raw === 'true') { badgeClass = 'success'; label = 'True'; }
                else if (raw === 'false') { badgeClass = 'danger'; label = 'False'; }
                else if (raw === '' || raw === 'n/a' || raw === 'na' || raw === 'none' || raw === 'null' || raw === 'undefined') { badgeClass = 'secondary'; label = 'N/A'; }
                return `<td><span class="badge bg-${badgeClass}">${escapeHTML(label)}</span></td>`;
            }
            if (h === 'complianceState') {
                const badgeClass = value.toLowerCase() === 'noncompliant' ? 'danger' : 'success';
                return `<td><span class="badge bg-${badgeClass}">${escapeHTML(value)}</span></td>`;
            }
            if (h === 'recommendationSeverity') {
                const badgeClass = value.toLowerCase() === 'high' ? 'danger' : value.toLowerCase() === 'medium' ? 'warning' : 'info';
                return `<td><span class="badge bg-${badgeClass}">${escapeHTML(value)}</span></td>`;
            }
            // Tier formatting for defender dataset
            if (datasetName === 'defender' && h === 'tier') {
                const badgeClass = value.toLowerCase() === 'standard' ? 'success' : value.toLowerCase() === 'free' ? 'warning' : 'secondary';
                return `<td><span class="badge bg-${badgeClass}">${escapeHTML(value)}</span></td>`;
            }
            // Value formatting for costs dataset (accounting format with 2 decimals)
            if (datasetName === 'costs' && h === 'value') {
                const numValue = parseFloat(value);
                if (!isNaN(numValue)) {
                    const formatted = numValue.toLocaleString('en-US', {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2
                    });
                    return `<td class="text-end">${formatted}</td>`;
                }
                return `<td class="text-end">${escapeHTML(value)}</td>`;
            }
            // SLA formatting for inventory dataset
            if (datasetName === 'inventory' && h === 'sla') {
                const lowerValue = value.toLowerCase().trim();
                // Empty or N/A = blank cell
                if (lowerValue === '' || lowerValue === 'n/a') {
                    return `<td></td>`;
                }
                // Explicit "None" = red
                if (lowerValue === 'none') {
                    return `<td><span class="badge bg-danger">None</span></td>`;
                }
                // Valid SLA value = green
                return `<td><span class="badge bg-success">${escapeHTML(value)}</span></td>`;
            }
            // Arc SQL dataset formatting
            if (datasetName === 'arcSQL') {
                // License Type formatting (SA/PAYG = green, unset = gray)
                if (h === 'license') {
                    const lowerValue = value.toLowerCase().trim();
                    if (lowerValue === '' || !value) return `<td></td>`;
                    let badgeClass = 'success';
                    if (lowerValue === 'unset') badgeClass = 'secondary';
                    return `<td><span class="badge bg-${badgeClass}">${escapeHTML(value)}</span></td>`;
                }
                // DPS Status formatting (OK = success, others = warning)
                if (h === 'dpsStatus') {
                    const lowerValue = value.toLowerCase().trim();
                    if (lowerValue === '' || !value) return `<td></td>`;
                    const badgeClass = lowerValue === 'ok' ? 'success' : (lowerValue === 'no data' ? 'secondary' : 'warning');
                    return `<td><span class="badge bg-${badgeClass}">${escapeHTML(value)}</span></td>`;
                }
                // TEL Status formatting (OK/__  = success, others = warning)
                if (h === 'telStatus') {
                    const lowerValue = value.toLowerCase().trim();
                    if (lowerValue === '' || !value) return `<td></td>`;
                    const badgeClass = (lowerValue === 'ok' || lowerValue === '__') ? 'success' : (lowerValue === 'no data' ? 'secondary' : 'warning');
                    const displayValue = lowerValue === '__' ? 'OK' : value;
                    return `<td><span class="badge bg-${badgeClass}">${escapeHTML(displayValue)}</span></td>`;
                }
                // Defender Status formatting (Protected = success, others = danger/gray)
                if (h === 'defenderStatus') {
                    const lowerValue = value.toLowerCase().trim();
                    if (lowerValue === '' || !value) return `<td></td>`;
                    let badgeClass = 'secondary';
                    if (lowerValue === 'protected') badgeClass = 'success';
                    else if (lowerValue === 'not protected') badgeClass = 'danger';
                    return `<td><span class="badge bg-${badgeClass}">${escapeHTML(value)}</span></td>`;
                }
            }
            // Render clipboard button for resourceId
            if (h === 'resourceId' && value) {
                return `<td><button class="btn btn-sm btn-outline-secondary copy-to-clipboard" data-clipboard="${escapeHTML(value)}" title="Copy Resource ID to clipboard">
                    <i class="bi bi-clipboard"></i>
                </button></td>`;
            }
            return `<td><span class="text-break">${escapeHTML(value)}</span></td>`;
        }).join('');
        html += '</tr>';
    });
    html += '</tbody>';

    table.innerHTML = html;
    console.log('Table rendered with', rows.length, 'rows');

    // Add filtering functionality with delay to ensure DOM is ready
    setTimeout(() => {
        try {
            addTableFilters();
        } catch (error) {
            console.warn('Filter setup failed:', error);
        }
    }, 100);
}

function renderPluginTable(rows, plugin) {
    const table = document.getElementById('data-table');
    if (rows.length === 0) {
        table.innerHTML = '<tbody><tr><td class="text-center text-muted py-4"><i class="bi bi-inbox me-2"></i>No data available</td></tr></tbody>';
        return;
    }

    const columns = plugin.columns || [];
    const readableHeaders = columns.map(c => c.name);

    // Build table with filter headers based on metadata
    let html = '<thead class="table-dark sticky-top">';
    html += '<tr>' + readableHeaders.map(h => `<th scope="col" class="fw-semibold">${h}</th>`).join('') + '</tr>';

    // Add filter row with dynamic filter types
    html += '<tr class="table-secondary">';
    html += columns.map((col, index) => {
        const filterType = col.filterType || 'none';
        if (filterType === 'none') {
            return `<th class="p-1"></th>`;
        } else if (filterType === 'dropdown') {
            // Get unique values for dropdown - use dataKey if available
            const dataKey = col.dataKey || col.name;
            const valueMap = new Map();
            rows.forEach(r => {
                const val = r[dataKey];
                if (val && val.trim()) {
                    const lowerKey = val.toLowerCase();
                    if (!valueMap.has(lowerKey)) {
                        valueMap.set(lowerKey, val);
                    }
                }
            });
            const uniqueValues = Array.from(valueMap.values()).sort((a, b) =>
                a.toLowerCase().localeCompare(b.toLowerCase())
            );

            return `<th class="p-1">
                <select class="form-select form-select-sm table-filter" data-column="${index}">
                    <option value="">All</option>
                    ${uniqueValues.map(v => `<option value="${escapeHTML(v)}">${escapeHTML(v)}</option>`).join('')}
                </select>
            </th>`;
        } else { // search
            return `<th class="p-1">
                <input type="text" class="form-control form-control-sm table-filter" data-column="${index}" placeholder="Filter...">
            </th>`;
        }
    }).join('');
    html += '</tr></thead>';

    // Build table body
    html += '<tbody>';
    rows.forEach((r, rowIndex) => {
        html += `<tr class="filterable-row ${rowIndex % 2 === 0 ? 'table-light' : ''}">`;
        html += columns.map(col => {
            // Use dataKey if available, otherwise fall back to name
            const dataKey = col.dataKey || col.name;
            const value = r[dataKey] || '';
            return `<td><span class="text-break">${escapeHTML(value)}</span></td>`;
        }).join('');
        html += '</tr>';
    });
    html += '</tbody>';

    table.innerHTML = html;
    console.log('Plugin table rendered with', rows.length, 'rows');

    // Add filtering functionality
    setTimeout(() => {
        try {
            addTableFilters();
        } catch (error) {
            console.warn('Filter setup failed:', error);
        }
    }, 100);
}

function addTableFilters() {
    const filters = document.querySelectorAll('.table-filter');
    console.log('Setting up', filters.length, 'table filters');

    filters.forEach(filter => {
        filter.addEventListener('input', function () {
            filterTable();
        });
        filter.addEventListener('change', function () {
            filterTable();
        });
    });

    // Initial count update
    updateRowCount();
}

function filterTable() {
    const filters = document.querySelectorAll('.table-filter');
    const rows = document.querySelectorAll('.filterable-row');

    rows.forEach(row => {
        let visible = true;
        const cells = row.querySelectorAll('td');

        for (const filter of filters) {
            const rawVal = filter.value;
            if (!rawVal) continue; // skip empty
            const filterVal = rawVal.toLowerCase().trim();
            const colAttr = filter.getAttribute('data-column');
            if (colAttr === null) continue;
            const colIndex = parseInt(colAttr, 10);
            const cellText = cells[colIndex] ? cells[colIndex].textContent.toLowerCase() : '';

            // Decide match strategy: exact for selects, substring for text inputs
            const isSelect = filter.tagName === 'SELECT';
            const match = isSelect ? (cellText === filterVal) : cellText.includes(filterVal);
            if (!match) { visible = false; break; }
        }

        row.style.display = visible ? '' : 'none';
    });

    updateRowCount();
}

function updateRowCount() {
    const allRows = document.querySelectorAll('.filterable-row');
    const visibleRows = document.querySelectorAll('.filterable-row:not([style*="display: none"])');
    const filters = document.querySelectorAll('.table-filter');

    let countDisplay = document.getElementById('row-count');
    if (!countDisplay) {
        countDisplay = document.createElement('div');
        countDisplay.id = 'row-count';
        const container = document.getElementById('table-controls-container');
        if (container) {
            container.appendChild(countDisplay);
        } else {
            console.warn('table-controls-container not found, trying fallback');
            const tableCard = document.querySelector('.table-responsive');
            if (tableCard && tableCard.parentElement) {
                tableCard.parentElement.insertBefore(countDisplay, tableCard);
            }
        }
    }
    // Show filtered state if any filter has a non-empty value, even if it matches all rows
    const hasActiveFilters = Array.from(filters).some(f => f.value && f.value.trim() !== '');
    const isFiltered = hasActiveFilters; // we deliberately ignore row count equality here

    // Check if we need to rebuild (to avoid destroying buttons during click)
    const currentlyFiltered = countDisplay.querySelector('[data-action="clear-filters"]') !== null;
    const needsRebuild = currentlyFiltered !== isFiltered;

    if (!countDisplay.innerHTML || needsRebuild) {
        countDisplay.innerHTML = `
            <div>
                <i class="bi bi-table me-1"></i>
                Showing <strong>${visibleRows.length}</strong> of <strong>${allRows.length}</strong> rows
            </div>
            <div class="d-flex align-items-center gap-2">
                ${isFiltered ? `
                    <span class="text-warning"><i class="bi bi-funnel me-1"></i>Filtered</span>
                    <button data-action="clear-filters" class="btn btn-sm btn-outline-secondary" title="Clear Filters">
                        <i class="bi bi-x-circle"></i>
                    </button>
                ` : ''}
                <button data-action="refresh-data" class="btn btn-sm btn-outline-primary" title="Refresh Data">
                    <i class="bi bi-arrow-clockwise"></i>
                </button>
            </div>
        `;
    } else {
        // Just update the counts without rebuilding HTML
        const counts = countDisplay.querySelectorAll('strong');
        if (counts.length >= 2) {
            counts[0].textContent = visibleRows.length;
            counts[1].textContent = allRows.length;
        }
    }
}

function clearAllTableFilters() {
    console.log('clearAllTableFilters called');

    const filters = document.querySelectorAll('.table-filter');

    // Remove all event listeners temporarily by replacing with clones
    const clearedFilters = [];
    filters.forEach(filter => {
        const clone = filter.cloneNode(true);
        clone.value = '';
        filter.parentNode.replaceChild(clone, filter);
        clearedFilters.push(clone);
    });

    // Show all rows
    const rows = document.querySelectorAll('.filterable-row');
    rows.forEach(row => row.style.display = '');

    // Now update the count (filters are already cleared so hasActiveFilters will be false)
    updateRowCount();

    // Re-attach event listeners to the new cloned filters
    clearedFilters.forEach(filter => {
        filter.addEventListener('input', function () {
            filterTable();
        });
        filter.addEventListener('change', function () {
            filterTable();
        });
    });

    console.log('Filters cleared successfully');
}

function escapeHTML(str) { return str.replace(/[&<>\"]/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c])); }
window.addEventListener('hashchange', navigate);

// Global event delegation for row count buttons
document.addEventListener('click', function (event) {
    const actionElement = event.target.closest('[data-action]');
    if (!actionElement) return;

    const action = actionElement.getAttribute('data-action');
    console.log('Button clicked with action:', action);

    event.preventDefault();
    event.stopPropagation();

    if (action === 'refresh-data') {
        console.log('Refreshing data...');
        navigate();
    } else if (action === 'clear-filters') {
        console.log('Clearing filters...');
        clearAllTableFilters();
    }
});

// Global event delegation for clipboard copy buttons
document.addEventListener('click', function (event) {
    const clipboardButton = event.target.closest('.copy-to-clipboard');
    if (!clipboardButton) return;

    event.preventDefault();
    event.stopPropagation();

    const textToCopy = clipboardButton.getAttribute('data-clipboard');
    if (!textToCopy) return;

    navigator.clipboard.writeText(textToCopy).then(() => {
        // Visual feedback: temporarily change icon to check mark
        const icon = clipboardButton.querySelector('i');
        const originalClass = icon.className;
        icon.className = 'bi bi-check-lg';
        clipboardButton.classList.remove('btn-outline-secondary');
        clipboardButton.classList.add('btn-success');

        setTimeout(() => {
            icon.className = originalClass;
            clipboardButton.classList.remove('btn-success');
            clipboardButton.classList.add('btn-outline-secondary');
        }, 1000);
    }).catch(err => {
        console.error('Failed to copy to clipboard:', err);
        // Visual feedback for error
        const icon = clipboardButton.querySelector('i');
        const originalClass = icon.className;
        icon.className = 'bi bi-x-lg';
        clipboardButton.classList.remove('btn-outline-secondary');
        clipboardButton.classList.add('btn-danger');

        setTimeout(() => {
            icon.className = originalClass;
            clipboardButton.classList.remove('btn-danger');
            clipboardButton.classList.add('btn-outline-secondary');
        }, 1000);
    });
});

initNav().then(navigate);
function makeReadable(str) {
    const mappings = {
        'subscriptionId': 'Subscription ID',
        'subscriptionName': 'Subscription Name',
        'resourceType': 'Resource Type',
        'resourceName': 'Resource Name',
        'resourceGroup': 'Resource Group',
        'resourceId': 'Resource ID',
        'impactedCount': 'Count',
        'impactedTotal': 'Total',
        'highImpact': 'High',
        'mediumImpact': 'Medium',
        'lowImpact': 'Low',
        'complianceState': 'Compliance State',
        'recommendationId': 'Recommendation ID',
        'numberOfImpactedResources': 'Impacted Resources',
        'azureServiceWellArchitected': 'Azure Service',
        'azureServiceCategoryWellArchitectedArea': 'Azure Service Category Well Architected Area',
        'azureServiceWellArchitectedTopic': 'Azure Service Well Architected Topic',
        'recommendationSource': 'Source',
        'bestPracticesGuidance': 'Best Practices',
        'readMore': 'Learn More',
        'azureArcServer': 'Azure Arc Server',
        'sqlInstance': 'SQL Instance',
        'vCores': 'VCores',
        'patchLevel': 'Patch Level',
        'dpsStatus': 'DPS Status',
        'telStatus': 'TEL Status',
        'defenderStatus': 'Defender Status'
    };
    return mappings[str] || str.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase()).trim();
}
function isValidUrl(string) {
    try {
        const url = new URL(string);
        return url.protocol === 'http:' || url.protocol === 'https:';
    } catch (_) {
        return false;
    }
}
