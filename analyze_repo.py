import subprocess
import os
import json

CATEGORIES = {
    "Application Code": ["*.go", "templates/", "static/"],
    "IaC": ["terraform/", "k8s/", "Dockerfile", "docker-compose.yml"],
    "Database": ["init.sql", "migrations/"],
    "CI/CD": [".github/"],
    "Documentation": ["*.md", "docs/", "LICENSE"],
    "Scripts": ["scripts/"],
    "Config": [".env", "go.mod", "go.sum", ".gitignore"]
}

def get_category(filename):
    for cat, patterns in CATEGORIES.items():
        for pattern in patterns:
            if pattern.endswith("/"):
                if filename.startswith(pattern) or ("/" + pattern) in filename:
                    return cat
            elif pattern.startswith("*"):
                if filename.endswith(pattern[1:]):
                    return cat
            else:
                if filename == pattern or filename.endswith("/" + pattern):
                    return cat
    return "Other"

def run_command(cmd):
    return subprocess.check_output(cmd, shell=True).decode('utf-8').strip()

def analyze_main():
    print("Analyzing main branch...")
    files = run_command("git ls-tree -r main --name-only").split('\n')
    stats = {}
    
    for f in files:
        if not f: continue
        try:
            content = run_command(f"git show main:'{f}'")
            lines = len(content.split('\n'))
            cat = get_category(f)
            stats[cat] = stats.get(cat, 0) + lines
        except Exception as e:
            print(f"Error reading {f}: {e}")
            
    return stats

def analyze_branches(branches):
    branch_stats = {}
    for branch in branches:
        print(f"Analyzing branch {branch}...")
        try:
            # numstat gives: added deleted filename
            diff = run_command(f"git diff --numstat main...{branch}")
            stats = {}
            for line in diff.split('\n'):
                if not line: continue
                parts = line.split()
                if len(parts) < 3: continue
                added = int(parts[0]) if parts[0] != '-' else 0
                # deleted = int(parts[1]) if parts[1] != '-' else 0
                filename = " ".join(parts[2:])
                
                cat = get_category(filename)
                stats[cat] = stats.get(cat, 0) + added
            branch_stats[branch] = stats
        except Exception as e:
            print(f"Error analyzing branch {branch}: {e}")
    return branch_stats

def main():
    main_stats = analyze_main()
    branches = ["feature/gke-base-deployment", "feature/ha-scalability-hardening", "feature/risk-mitigation"]
    branch_stats = analyze_branches(branches)
    
    report = {
        "main": main_stats,
        "branches": branch_stats
    }
    
    print(json.dumps(report, indent=2))

if __name__ == "__main__":
    main()
