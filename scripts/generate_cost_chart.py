import matplotlib.pyplot as plt
import numpy as np
import os

# Ensure the output directory exists
output_dir = 'docs/images'
os.makedirs(output_dir, exist_ok=True)

milestones = ["M00", "M01", "M02\n(Nov 20)", "M03", "M04", "M05", "M06", "M07", "M08", "M09", "M10", "M11"]
cloud_sql = np.array([0, 0, 1.0, 6.2, 6.2, 6.2, 6.2, 6.2, 6.2, 6.2, 6.2, 6.2])
gke_nodes = np.array([0, 0, 4.0, 7.2, 7.2, 7.2, 7.2, 7.2, 7.2, 7.2, 7.2, 7.2])
other = np.array([0, 0, 1.0, 1.6, 1.6, 1.6, 1.6, 2.1, 2.6, 3.1, 3.4, 3.6])

ind = np.arange(len(milestones))
width = 0.65

fig, ax = plt.subplots(figsize=(12, 7))

# Plot stacked bars
p1 = ax.bar(ind, cloud_sql, width, color='#9b59b6', label='Cloud SQL (HA + Replica)', zorder=3)
p2 = ax.bar(ind, gke_nodes, width, bottom=cloud_sql, color='#3498db', label='GKE & Compute', zorder=3)
p3 = ax.bar(ind, other, width, bottom=cloud_sql+gke_nodes, color='#e67e22', label='Observability & Net', zorder=3)

# Formatting
ax.set_ylabel('Daily Cost ($)', fontsize=12, fontweight='bold')
ax.set_title('Daily Cost Evolution by Milestone', fontsize=16, fontweight='bold', pad=20)
ax.set_xticks(ind)
ax.set_xticklabels(milestones, fontsize=10)
ax.set_ylim(0, 20)
ax.legend(loc='upper left', frameon=True, fontsize=10)
ax.grid(axis='y', linestyle='--', alpha=0.5, zorder=0)

# Remove top and right spines
ax.spines['top'].set_visible(False)
ax.spines['right'].set_visible(False)

# Add total labels
totals = cloud_sql + gke_nodes + other
for i, v in enumerate(totals):
    if v > 0:
        ax.text(i, v + 0.3, f"${v:.2f}", ha='center', fontweight='bold', fontsize=10)

plt.tight_layout()
output_path = os.path.join(output_dir, 'daily_cost_chart.png')
plt.savefig(output_path, dpi=300)
print(f"Chart saved to {output_path}")
