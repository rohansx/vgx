import * as vscode from 'vscode';

// AI code detection patterns (simplified TypeScript port)
const AI_PATTERNS = [
    { name: 'try_catch', pattern: /try\s*\{[^}]+\}\s*catch\s*\(\s*(?:error|err|e)\s*(?::\s*\w+)?\s*\)\s*\{[^}]*(?:console\.(?:error|log)|throw)[^}]*\}/g, weight: 0.15 },
    { name: 'async_fetch', pattern: /async\s+(?:function\s+)?\w+\s*\([^)]*\).*\{[^}]*await\s+fetch/g, weight: 0.10 },
    { name: 'arrow_types', pattern: /const\s+\w+\s*=\s*(?:async\s*)?\([^)]*:\s*\w+[^)]*\)\s*(?::\s*\w+(?:<[^>]+>)?)?\s*=>/g, weight: 0.10 },
    { name: 'use_effect', pattern: /useEffect\s*\(\s*\(\s*\)\s*=>\s*\{[^}]+\}\s*,\s*\[[^\]]*\]\s*\)/g, weight: 0.08 },
    { name: 'use_state', pattern: /const\s*\[\s*\w+\s*,\s*set[A-Z]\w+\s*\]\s*=\s*useState/g, weight: 0.08 },
    { name: 'jsdoc', pattern: /\/\*\*\s*\n(?:\s*\*\s*@\w+[^\n]*\n)+\s*\*\//g, weight: 0.10 },
];

interface DetectionResult {
    confidence: number;
    isAI: boolean;
    patterns: string[];
}

let aiDecorations: vscode.TextEditorDecorationType;
let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
    console.log('VGX extension activated');

    // Create decoration type for AI code highlighting
    aiDecorations = vscode.window.createTextEditorDecorationType({
        backgroundColor: 'rgba(255, 193, 7, 0.1)',
        border: '1px solid rgba(255, 193, 7, 0.3)',
        borderRadius: '2px',
        overviewRulerColor: 'rgba(255, 193, 7, 0.8)',
        overviewRulerLane: vscode.OverviewRulerLane.Right,
    });

    // Create status bar item
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.command = 'vgx.detectFile';
    context.subscriptions.push(statusBarItem);

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('vgx.detectFile', detectCurrentFile),
        vscode.commands.registerCommand('vgx.detectWorkspace', detectWorkspace),
        vscode.commands.registerCommand('vgx.scanFile', scanCurrentFile)
    );

    // Update on document change
    context.subscriptions.push(
        vscode.window.onDidChangeActiveTextEditor(editor => {
            if (editor) {
                updateDecorations(editor);
            }
        }),
        vscode.workspace.onDidChangeTextDocument(event => {
            const editor = vscode.window.activeTextEditor;
            if (editor && event.document === editor.document) {
                updateDecorations(editor);
            }
        })
    );

    // Initial update
    if (vscode.window.activeTextEditor) {
        updateDecorations(vscode.window.activeTextEditor);
    }
}

function analyzeCode(text: string): DetectionResult {
    let totalWeight = 0;
    const matchedPatterns: string[] = [];

    for (const { name, pattern, weight } of AI_PATTERNS) {
        const matches = text.match(pattern);
        if (matches && matches.length > 0) {
            totalWeight += weight * matches.length;
            matchedPatterns.push(name);
        }
    }

    // Analyze naming consistency
    const identifiers = text.match(/\b([a-z][a-zA-Z0-9_]*)\b/g) || [];
    if (identifiers.length > 5) {
        const camelCase = identifiers.filter(i => /^[a-z]+([A-Z][a-z]*)*$/.test(i)).length;
        const consistency = camelCase / identifiers.length;
        if (consistency > 0.8) {
            totalWeight += 0.15;
        }
    }

    // Cap at 1.0
    const confidence = Math.min(totalWeight, 1.0);
    const config = vscode.workspace.getConfiguration('vgx');
    const threshold = (config.get('aiThreshold') as number || 70) / 100;

    return {
        confidence,
        isAI: confidence >= threshold,
        patterns: matchedPatterns
    };
}

function updateDecorations(editor: vscode.TextEditor) {
    const config = vscode.workspace.getConfiguration('vgx');
    if (!config.get('highlightAICode')) {
        editor.setDecorations(aiDecorations, []);
        statusBarItem.hide();
        return;
    }

    const text = editor.document.getText();
    const result = analyzeCode(text);

    // Update status bar
    const confidencePercent = Math.round(result.confidence * 100);
    if (result.isAI) {
        statusBarItem.text = `$(robot) ${confidencePercent}% AI`;
        statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground');
        statusBarItem.tooltip = `AI-generated code detected (${result.patterns.join(', ')})`;
    } else {
        statusBarItem.text = `$(check) ${confidencePercent}% AI`;
        statusBarItem.backgroundColor = undefined;
        statusBarItem.tooltip = 'Human-written code';
    }
    statusBarItem.show();

    // Highlight patterns if AI detected
    if (result.isAI && config.get('showInlineConfidence')) {
        const decorations: vscode.DecorationOptions[] = [];
        
        for (const { pattern } of AI_PATTERNS) {
            let match;
            const regex = new RegExp(pattern.source, 'g');
            while ((match = regex.exec(text)) !== null) {
                const startPos = editor.document.positionAt(match.index);
                const endPos = editor.document.positionAt(match.index + match[0].length);
                decorations.push({
                    range: new vscode.Range(startPos, endPos),
                    hoverMessage: 'Likely AI-generated pattern'
                });
            }
        }
        
        editor.setDecorations(aiDecorations, decorations);
    } else {
        editor.setDecorations(aiDecorations, []);
    }
}

async function detectCurrentFile() {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No file open');
        return;
    }

    const text = editor.document.getText();
    const result = analyzeCode(text);
    const confidencePercent = Math.round(result.confidence * 100);

    if (result.isAI) {
        vscode.window.showWarningMessage(
            `🤖 AI-generated code detected: ${confidencePercent}% confidence\n` +
            `Patterns: ${result.patterns.join(', ')}`
        );
    } else {
        vscode.window.showInformationMessage(
            `✓ Human-written code: ${confidencePercent}% AI confidence`
        );
    }
}

async function detectWorkspace() {
    const files = await vscode.workspace.findFiles('**/*.{ts,tsx,js,jsx,py,go}', '**/node_modules/**');
    
    let aiCount = 0;
    let totalFiles = 0;

    await vscode.window.withProgress({
        location: vscode.ProgressLocation.Notification,
        title: 'VGX: Scanning workspace...',
        cancellable: true
    }, async (progress, token) => {
        for (const file of files) {
            if (token.isCancellationRequested) break;
            
            const doc = await vscode.workspace.openTextDocument(file);
            const result = analyzeCode(doc.getText());
            
            if (result.isAI) aiCount++;
            totalFiles++;
            
            progress.report({ 
                message: `${totalFiles}/${files.length} files`,
                increment: 100 / files.length
            });
        }
    });

    vscode.window.showInformationMessage(
        `VGX: ${aiCount}/${totalFiles} files detected as AI-generated`
    );
}

async function scanCurrentFile() {
    vscode.window.showInformationMessage('Security scan coming soon!');
}

export function deactivate() {
    if (aiDecorations) {
        aiDecorations.dispose();
    }
}
