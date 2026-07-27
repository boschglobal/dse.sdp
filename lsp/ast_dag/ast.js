"use strict";

const model_colors = [
    "#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", 
    "#98D8C8", "#F7DC6F", "#BB8FCE", "#85C1E2",
    "#F8B88B", "#A9DFBF", "#F5B7B1", "#D7BDE2",
    "#82E0AA", "#F9E79F", "#FADBD8", "#D5F4E6",
    "#EBDEF0", "#A3E4D7", "#F4ECF7", "#D2EE9F"
];

let channel_colors = [
    "#FFDDC1", "#FFD700", "#FFDAA5", "#FFC3A0",
    "#FFB6C1", "#C5E1A5", "#B3E5FC", "#D1C4E9",
    "#F8BBD0", "#F5DEB3", "#F0E68C", "#E0FFFF",
    "#D3E0EA", "#E6E6FA", "#FAECCC", "#FFFACD",
];

let modelColorMap = {};

const rect_height = 40;
const rect_width = 110;
const horizontal_rect_height = 10;
const horizontal_rect_width = 400;
const vertical_rounded_rect_width = 20;
const vertical_rounded_rect_height = 70;
const horizontal_rounded_rect_height = 20;
const horizontal_rounded_rect_width = 110;

function getRandomColor() {
    const random_color = channel_colors[Math.floor(Math.random() * channel_colors.length)];
    channel_colors = channel_colors.filter(color => color !== random_color);
    return random_color;
}

function assignModelColors(graph) {
    modelColorMap = {};
    let colorIndex = 0;
    graph.nodes.forEach(node => {
        if (node.type === 'rect') {
            modelColorMap[node.id] = model_colors[colorIndex % model_colors.length];
            colorIndex++;
        }
    });
}

function findSourceModel(nodeId, graph) {
    const node = graph.nodes.find(n => n.id === nodeId);
    if (node && node.type === 'rect') {
        return nodeId;
    }
    
    for (let link of graph.links) {
        if (link.target.id === nodeId) {
            const sourceNode = graph.nodes.find(n => n.id === link.source.id);
            if (sourceNode && sourceNode.type === 'rect') {
                return link.source.id;
            }
            const parentModel = findSourceModel(link.source.id, graph);
            if (parentModel) return parentModel;
        }
    }
    return null;
}

function getNodeBounds(node) {
    let width = 0, height = 0;
    if (node.type === 'rect') {
        width = rect_width;
        height = rect_height;
    } else if (node.type === 'horizontal_rect') {
        width = horizontal_rect_width;
        height = horizontal_rect_height;
    } else if (node.type === 'vertical_rounded_rect') {
        width = vertical_rounded_rect_width;
        height = vertical_rounded_rect_height;
    } else if (node.type === 'horizontal_rounded_rect') {
        width = horizontal_rounded_rect_width;
        height = horizontal_rounded_rect_height;
    }
    
    return {
        x: node.x - width / 2,
        y: node.y - height / 2,
        x2: node.x + width / 2,
        y2: node.y + height / 2,
        width: width,
        height: height
    };
}

function truncateText(text, type, graph, id) {
    const maxTextLength = 13;
    if (text.length > maxTextLength) {
        text = text.slice(0, maxTextLength) + '...';
    }

    if (type === 'rect') {
        return text;
    } else if (type === 'vertical_rounded_rect' || type === 'horizontal_rounded_rect') {
        return `[ ${text} ]`;
    } else if (type === 'horizontal_rect') {
        let parent = "";
        for (let link of graph.links) {
            if (id === link.target.id) {
                parent = link.source.name;
            }
        }
        return `${text} [ ${parent} ]`;
    }
    return text;
}

function plotTree(graph) {
    assignModelColors(graph);
    
    const svg = d3.select('svg')
        .attr('width', window.innerWidth)
        .attr('height', window.innerHeight);

    svg.append('defs').append('marker')
        .attr('id', 'arrow')
        .attr('viewBox', '0 -5 10 10')
        .attr('refX', 8)
        .attr('refY', 0)
        .attr('markerWidth', 5)
        .attr('markerHeight', 3)
        .attr('orient', 'auto')
        .append('path')
        .attr('d', 'M0,-5L10,0L0,5')
        .attr('fill', '#000000')
        .style('z-index', 9999);

    const gridSpacing = 200;
    const simulation = d3.forceSimulation(graph.nodes)
        .force("link", d3.forceLink(graph.links).id(d => d.id).distance(gridSpacing))
        .force("charge", d3.forceManyBody().strength(-(gridSpacing * 2)))
        .force("x", d3.forceX(d => Math.round(d.x / gridSpacing) * gridSpacing).strength(1))
        .force("y", d3.forceY(d => Math.round(d.y / gridSpacing) * gridSpacing).strength(1))
        .force("collide", d3.forceCollide(gridSpacing * 0.8)) // Reduce overlap risk
        .alphaDecay(0.08);

    const node = svg.selectAll('.node')
        .data(graph.nodes)
        .enter().append('g')
        .attr('class', 'node')
        .attr('transform', d => `translate(${d.x || 0}, ${d.y || 0})`);

    node.each(function (d) {
        if (d.type === 'rect') {
            const nodeSelection = d3.select(this).append('rect')
                .attr("y", 0)
                .attr("width", rect_width)
                .attr("height", rect_height)
                .attr("fill", "#55caec");

        } else if (d.type === 'vertical_rounded_rect') {
            d3.select(this).append('rect')
                .attr("width", vertical_rounded_rect_width)
                .attr("height", vertical_rounded_rect_height)
                .attr("rx", 10) // Rounded corners for vertical_rounded_rect
                .attr("ry", 10)
                .attr("fill", getRandomColor());

        } else if (d.type === 'horizontal_rounded_rect') {
            d3.select(this).append('rect')
                .attr("width", horizontal_rounded_rect_width)
                .attr("height", horizontal_rounded_rect_height)
                .attr("rx", 10) // Rounded corners for horizontal_rounded_rect
                .attr("ry", 10)
                .attr("fill", getRandomColor());

        } else if (d.type === 'horizontal_rect') {
            d3.select(this).append('rect')
                .attr("width", horizontal_rect_width)
                .attr("height", horizontal_rect_height)
                .attr("fill", getRandomColor());
        }
    });

    node.append('text')
        .attr('dx', function (d) {
            if (d.type === 'rect') {
                return 55;
            } else if (d.type === 'vertical_rounded_rect') {
                return 10;
            }
            else if (d.type === 'horizontal_rounded_rect') {
                return 55;
            }
            else if (d.type === 'horizontal_rect') {
                return 200;
            }
        })
        .attr('dy', function (d) {
            if (d.type === 'rect') {
                return 25;
            } else if (d.type === 'vertical_rounded_rect') {
                return -10;
            } else if (d.type === 'horizontal_rounded_rect') {
                return -12;
            } else if (d.type === 'horizontal_rect') {
                return -15;
            }
        })
        .attr('text-anchor', 'middle')
        .text(function (d) {
            return truncateText(d.name, d.type, graph, d.id);
        });


    const link = svg.selectAll('.link')
        .data(graph.links)
        .enter().append('path')
        .attr('class', 'link')
        .attr('fill', 'none')
        .attr('stroke', function(d) {
            const sourceModel = findSourceModel(d.source.id, graph);
            return sourceModel && modelColorMap[sourceModel] ? modelColorMap[sourceModel] : '#999';
        })
        .attr('stroke-width', 2.5)
        .attr('stroke-opacity', 0.7)
        .attr('marker-end', 'url(#arrow)');

    function closestPoint(source, target) {
        let closest = { x: source.x, y: source.y };

        let width = 0;
        let height = 0;
        if (source.type === 'rect') {
            width = rect_width;
            height = rect_height;
        } else if (source.type === 'horizontal_rect') {
            width = horizontal_rect_width;
            height = horizontal_rect_height;
        } else if (source.type === 'vertical_rounded_rect') {
            width = vertical_rounded_rect_width;
            height = vertical_rounded_rect_height;
        } else if (source.type === 'horizontal_rounded_rect') {
            width = horizontal_rounded_rect_width;
            height = horizontal_rounded_rect_height;
        }

        let borderPoints = [];
        if (source.type === 'vertical_rounded_rect') {
            borderPoints = [];
            borderPoints.push({ x: source.x + (width / 2), y: source.y }); // top border center
            borderPoints.push({ x: source.x + (width / 2), y: source.y + height }); // bottom border center
            borderPoints.push({ x: source.x, y: source.y + (height / 2) }); // left border center
            borderPoints.push({ x: source.x + width, y: source.y + (height / 2) }); // right border center
        } else if (source.type === 'horizontal_rounded_rect') {
            borderPoints = [];
            borderPoints.push({ x: source.x + (width / 2), y: source.y }); // top border center
            borderPoints.push({ x: source.x + (width / 2), y: source.y + height }); // bottom border center
            borderPoints.push({ x: source.x, y: source.y + (height / 2) }); // left border center
            borderPoints.push({ x: source.x, y: source.y + (height / 2) - 5 }); // left border below center
            borderPoints.push({ x: source.x, y: source.y + (height / 2) + 5 }); // left border above center
            borderPoints.push({ x: source.x + width, y: source.y + (height / 2) }); // right border center
            borderPoints.push({ x: source.x + width, y: source.y + (height / 2) + 5 }); // right border above center
            borderPoints.push({ x: source.x + width, y: source.y + (height / 2) - 5 }); // right border elow center
        } else if (source.type === 'rect') {
            borderPoints = [];
            borderPoints.push({ x: source.x + (width / 2), y: source.y }); // top border center
            borderPoints.push({ x: source.x + (width / 2), y: source.y + height }); // bottom border center
            borderPoints.push({ x: source.x, y: source.y + (height / 2) }); // left border center
            borderPoints.push({ x: source.x + width, y: source.y + (height / 2) }); // right border center
        } else if (source.type === 'horizontal_rect') {
            borderPoints = [];
            borderPoints.push({ x: source.x + (width / 2), y: source.y }); // top border center
            borderPoints.push({ x: source.x + (width / 2), y: source.y + height }); // bottom border center
            borderPoints.push({ x: source.x, y: source.y + (height / 2) }); // left border center
            borderPoints.push({ x: source.x + width, y: source.y + (height / 2) }); // right border center
        }

        closest = borderPoints.reduce((prev, curr) =>
            Math.hypot(curr.x - target.x, curr.y - target.y) <
                Math.hypot(prev.x - target.x, prev.y - target.y)
                ? curr : prev
        );
        return closest;
    }


    const linkOffsets = new Map();
    
    function getRoutingOffset(sourceId, targetId) {
        const key = `${sourceId}-${targetId}`;
        let offset = linkOffsets.get(key);
        if (offset === undefined) {
            offset = (linkOffsets.size % 5) - 2;  // Range: -2, -1, 0, 1, 2
            linkOffsets.set(key, offset);
        }
        return offset * 25;  // Spread lines by 25px each
    }
    
    function lineSegmentIntersectsBox(p1, p2, bounds, padding) {
        const x1 = bounds.x - padding;
        const y1 = bounds.y - padding;
        const x2 = bounds.x2 + padding;
        const y2 = bounds.y2 + padding;
        
        if ((p1.x >= x1 && p1.x <= x2 && p1.y >= y1 && p1.y <= y2) ||
            (p2.x >= x1 && p2.x <= x2 && p2.y >= y1 && p2.y <= y2)) {
            return true;
        }
        
        const dx = p2.x - p1.x;
        const dy = p2.y - p1.y;
        
        if (Math.abs(dx) > 0.1) {
            for (let t of [(x1 - p1.x) / dx, (x2 - p1.x) / dx]) {
                if (t >= 0 && t <= 1) {
                    const y = p1.y + t * dy;
                    if (y >= y1 && y <= y2) return true;
                }
            }
        }
        
        if (Math.abs(dy) > 0.1) {
            for (let t of [(y1 - p1.y) / dy, (y2 - p1.y) / dy]) {
                if (t >= 0 && t <= 1) {
                    const x = p1.x + t * dx;
                    if (x >= x1 && x <= x2) return true;
                }
            }
        }
        
        return false;
    }
    
    function checkPathCollisions(pathSegments, sourceModel, targetId, allNodes) {
        // Check if path collides with any nodes
        let collisionCount = 0;
        
        for (let i = 0; i < pathSegments.length - 1; i++) {
            const p1 = pathSegments[i];
            const p2 = pathSegments[i + 1];
            
            for (let node of allNodes) {
                const nodeModel = findSourceModel(node.id, graph);
                // Skip source model and target node
                if (nodeModel === sourceModel || node.id === targetId) continue;
                
                const bounds = getNodeBounds(node);
                if (lineSegmentIntersectsBox(p1, p2, bounds, 40)) {
                    collisionCount++;
                }
            }
        }
        
        return collisionCount;
    }
    
    function generateSmartPath(source, target, sourceModel, allNodes) {
        const start = closestPoint(source, target);
        const end = closestPoint(target, source);

        if (!start || !end) return "";

        const offset = getRoutingOffset(source.id, target.id);
        const dx = end.x - start.x;
        const dy = end.y - start.y;
        const dist = Math.sqrt(dx * dx + dy * dy);
        
        // Generate multiple candidate paths
        const candidates = [];
        
        // Strategy 1: Horizontal first (with offset)
        {
            const mid1X = start.x + dx * 0.4;
            const mid1Y = start.y + offset;
            const path = `M ${start.x} ${start.y} L ${mid1X} ${mid1Y} L ${end.x} ${end.y}`;
            const points = [{x: start.x, y: start.y}, {x: mid1X, y: mid1Y}, {x: end.x, y: end.y}];
            const collisions = checkPathCollisions(points, sourceModel, target.id, allNodes);
            candidates.push({path, collisions, offset: offset});
        }
        
        // Strategy 2: Vertical first (with offset)
        {
            const mid2X = start.x + offset;
            const mid2Y = start.y + dy * 0.4;
            const path = `M ${start.x} ${start.y} L ${mid2X} ${mid2Y} L ${end.x} ${end.y}`;
            const points = [{x: start.x, y: start.y}, {x: mid2X, y: mid2Y}, {x: end.x, y: end.y}];
            const collisions = checkPathCollisions(points, sourceModel, target.id, allNodes);
            candidates.push({path, collisions, offset: offset});
        }
        
        // Strategy 3: Different horizontal split point
        {
            const mid3X = start.x + dx * 0.6;
            const mid3Y = start.y + offset;
            const path = `M ${start.x} ${start.y} L ${mid3X} ${mid3Y} L ${end.x} ${end.y}`;
            const points = [{x: start.x, y: start.y}, {x: mid3X, y: mid3Y}, {x: end.x, y: end.y}];
            const collisions = checkPathCollisions(points, sourceModel, target.id, allNodes);
            candidates.push({path, collisions, offset: offset});
        }
        
        // Strategy 4: Larger offset horizontal
        {
            const largeOffset = offset * 1.5;
            const mid4X = start.x + dx * 0.5;
            const mid4Y = start.y + largeOffset;
            const path = `M ${start.x} ${start.y} L ${mid4X} ${mid4Y} L ${end.x} ${end.y}`;
            const points = [{x: start.x, y: start.y}, {x: mid4X, y: mid4Y}, {x: end.x, y: end.y}];
            const collisions = checkPathCollisions(points, sourceModel, target.id, allNodes);
            candidates.push({path, collisions, offset: largeOffset});
        }
        
        // Strategy 5: Larger offset vertical
        {
            const largeOffset = offset * 1.5;
            const mid5X = start.x + largeOffset;
            const mid5Y = start.y + dy * 0.5;
            const path = `M ${start.x} ${start.y} L ${mid5X} ${mid5Y} L ${end.x} ${end.y}`;
            const points = [{x: start.x, y: start.y}, {x: mid5X, y: mid5Y}, {x: end.x, y: end.y}];
            const collisions = checkPathCollisions(points, sourceModel, target.id, allNodes);
            candidates.push({path, collisions, offset: largeOffset});
        }
        
        // Select best path: prefer fewer collisions
        candidates.sort((a, b) => a.collisions - b.collisions);
        return candidates[0].path;
    }

    simulation.on('tick', () => {
        adjustSVGSize(svg, graph);

        link.attr('d', function (d) {
            const sourceModel = findSourceModel(d.source.id, graph);
            return generateSmartPath(d.source, d.target, sourceModel, graph.nodes);
        });

        node.attr('transform', d => `translate(${d.x}, ${d.y})`);
    });

}

function updateSVGSize() {
    (async () => {
        const data = await loadLocalJSONFile();
        plotTree(data)
    })();
}
window.addEventListener('resize', updateSVGSize);
updateSVGSize();

async function loadLocalJSONFile() {
    try {
        const url = document.currentScript?.hasAttribute("codespace_url") ? `${document.currentScript.getAttribute("codespace_url")}?t=${new Date().getTime()}` : `http://127.0.0.1:3001/input.json?t=${new Date().getTime()}`;
        const response = await fetch(url, {
            method: 'GET',
            cache: "no-store",
            priority: "high",
            headers: {
                'Content-Type': 'application/json',
            }
        });

        if (!response.ok) {
            throw new Error('Network response was not ok');
        }

        const jsonData = await response.json();
        plotTree(jsonData);
        return jsonData;
    } catch (error) {
        console.error('Error loading the JSON:', error);
    }
}

function adjustSVGSize(svg, graph) {
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;

    graph.nodes.forEach(node => {
        minX = Math.min(minX, node.x);
        minY = Math.min(minY, node.y);
        maxX = Math.max(maxX, node.x);
        maxY = Math.max(maxY, node.y);
    });

    const padding = 100; // Space around the graph
    const width = maxX - minX + 10 * padding;
    const height = maxY - minY + 2 * padding;

    svg.attr('width', width).attr('height', height)
        .attr('viewBox', `${minX - padding} ${minY - padding} ${width} ${height}`);
}