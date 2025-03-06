"use strict";

let channel_colors = [
    "#FFDDC1",  "#FFD700", "#FFDAA5", "#FFC3A0", 
    "#FFB6C1", "#C5E1A5", "#B3E5FC", "#D1C4E9", 
    "#F8BBD0", "#F5DEB3", "#F0E68C", "#E0FFFF", 
    "#D3E0EA", "#E6E6FA", "#FAECCC", "#FFFACD", 
];

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

// function getModelCount(node) {
//     const count = node.nodes && Array.isArray(node.nodes) ? node.nodes.length : 0;
//     const width = window.innerWidth;
//     const height = window.innerHeight;
//     return {
//         count,
//         height,
//         width
//     };
// }

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
    // const graph_data = getModelCount(graph);
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
        .force("charge", d3.forceManyBody().strength(-(gridSpacing*2)))
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
        .attr('stroke', '#999')
        .attr('stroke-width', 2)
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


    simulation.on('tick', () => {
        adjustSVGSize(svg, graph);

        link.attr('d', function (d) {
            const start = closestPoint(d.source, d.target);
            const end = closestPoint(d.target, d.source);

            if (!start || !end) return "";

            let midX = (start.x + end.x) / 2;
            let midY = (start.y + end.y) / 2;

            let path = `M ${start.x} ${start.y} `;

            if (Math.abs(start.x - end.x) > Math.abs(start.y - end.y)) {
                // Horizontal first, then vertical
                path += `L ${midX} ${start.y} L ${midX} ${end.y} L ${end.x} ${end.y}`;
            } else {
                // Vertical first, then horizontal
                path += `L ${start.x} ${midY} L ${end.x} ${midY} L ${end.x} ${end.y}`;
            }

            return path;
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
        const url = `http://127.0.0.1:3001/input.json?t=${new Date().getTime()}`;
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
        console.log(jsonData);
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