import { useEffect, useRef } from 'react';
import * as THREE from 'three';

export function Logistics3D() {
  const containerRef = useRef(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // --- Scene Setup ---
    const scene = new THREE.Scene();
    scene.fog = new THREE.FogExp2(0xfafafa, 0.015);

    // --- Camera Setup ---
    const camera = new THREE.PerspectiveCamera(
      45,
      container.clientWidth / container.clientHeight,
      0.1,
      1000
    );
    camera.position.set(15, 12, 20);
    camera.lookAt(0, 1, 0);

    // --- Renderer Setup ---
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    container.appendChild(renderer.domElement);

    // --- Lights ---
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.85);
    scene.add(ambientLight);

    const dirLight = new THREE.DirectionalLight(0xffffff, 1.5);
    dirLight.position.set(10, 20, 10);
    dirLight.castShadow = true;
    dirLight.shadow.mapSize.width = 1024;
    dirLight.shadow.mapSize.height = 1024;
    dirLight.shadow.bias = -0.001;
    scene.add(dirLight);

    const pointLight1 = new THREE.PointLight(0x5a4fcf, 1.5, 20);
    pointLight1.position.set(-5, 3, -5);
    scene.add(pointLight1);

    const pointLight2 = new THREE.PointLight(0x00bfa5, 1.5, 20);
    pointLight2.position.set(5, 2, 5);
    scene.add(pointLight2);

    // --- Ground Grid & Sea ---
    const gridHelper = new THREE.GridHelper(60, 60, 0x00bfa5, 0xe2e8f0);
    gridHelper.position.y = -0.01;
    scene.add(gridHelper);

    const oceanGeo = new THREE.PlaneGeometry(100, 100);
    const oceanMat = new THREE.MeshStandardMaterial({
      color: 0xe6f4f1,
      roughness: 0.3,
      metalness: 0.1,
      transparent: true,
      opacity: 0.8,
    });
    const ocean = new THREE.Mesh(oceanGeo, oceanMat);
    ocean.rotation.x = -Math.PI / 2;
    ocean.position.y = -0.1;
    ocean.receiveShadow = true;
    scene.add(ocean);

    // --- Create 3D Cargo Ship ---
    const shipGroup = new THREE.Group();

    // Hull
    const hullGeo = new THREE.BoxGeometry(6, 0.8, 1.5);
    const hullMat = new THREE.MeshStandardMaterial({ color: 0x1e293b, roughness: 0.3, metalness: 0.7 });
    const hull = new THREE.Mesh(hullGeo, hullMat);
    hull.castShadow = true;
    hull.receiveShadow = true;
    shipGroup.add(hull);

    // Bow (Front Point)
    const bowGeo = new THREE.ConeGeometry(0.75, 1.5, 4);
    const bow = new THREE.Mesh(bowGeo, hullMat);
    bow.rotation.z = -Math.PI / 2;
    bow.position.set(3, 0, 0);
    bow.scale.set(1, 1, 0.5);
    bow.castShadow = true;
    shipGroup.add(bow);

    // Bridge/Cabin (Back Structure)
    const cabinGeo = new THREE.BoxGeometry(1.2, 1.4, 1.2);
    const cabinMat = new THREE.MeshStandardMaterial({ color: 0xe2e8f0, roughness: 0.2 });
    const cabin = new THREE.Mesh(cabinGeo, cabinMat);
    cabin.position.set(-2, 0.7, 0);
    cabin.castShadow = true;
    shipGroup.add(cabin);

    const chimneyGeo = new THREE.CylinderGeometry(0.15, 0.15, 0.8);
    const chimneyMat = new THREE.MeshStandardMaterial({ color: 0xef4444 });
    const chimney = new THREE.Mesh(chimneyGeo, chimneyMat);
    chimney.position.set(-2.2, 1.6, 0);
    shipGroup.add(chimney);

    // Cargo Containers (Stacked)
    const containerColors = [0x00bfa5, 0x5a4fcf, 0x3b82f6, 0xf59e0b];
    const containerGroup = new THREE.Group();

    for (let x = -0.8; x <= 1.6; x += 0.8) {
      for (let z = -0.4; z <= 0.4; z += 0.8) {
        for (let y = 0.5; y <= 0.9; y += 0.4) {
          if (Math.random() > 0.15) { // some missing containers for realism
            const boxGeo = new THREE.BoxGeometry(0.7, 0.35, 0.7);
            const boxMat = new THREE.MeshStandardMaterial({
              color: containerColors[Math.floor(Math.random() * containerColors.length)],
              roughness: 0.5,
              metalness: 0.1
            });
            const containerBox = new THREE.Mesh(boxGeo, boxMat);
            containerBox.position.set(x, y, z);
            containerBox.castShadow = true;
            containerBox.receiveShadow = true;
            containerGroup.add(containerBox);
          }
        }
      }
    }
    shipGroup.add(containerGroup);
    shipGroup.position.set(4, 0.4, 2);
    shipGroup.scale.set(0.9, 0.9, 0.9);
    scene.add(shipGroup);

    // --- Create 3D Truck & Port Terminal ---
    const portGroup = new THREE.Group();
    portGroup.position.set(-6, 0, 4);

    // Dock platform
    const dockGeo = new THREE.BoxGeometry(5, 0.6, 4);
    const dockMat = new THREE.MeshStandardMaterial({ color: 0x334155, roughness: 0.8 });
    const dock = new THREE.Mesh(dockGeo, dockMat);
    dock.position.y = 0.3;
    dock.receiveShadow = true;
    portGroup.add(dock);

    // Crane
    const craneGroup = new THREE.Group();
    craneGroup.position.set(1.5, 0.6, 0);
    const pillarGeo = new THREE.CylinderGeometry(0.15, 0.15, 3.5);
    const pillarMat = new THREE.MeshStandardMaterial({ color: 0xf59e0b, metalness: 0.8 });
    const pillar = new THREE.Mesh(pillarGeo, pillarMat);
    pillar.position.y = 1.75;
    craneGroup.add(pillar);

    const armGeo = new THREE.BoxGeometry(3, 0.25, 0.4);
    const arm = new THREE.Mesh(armGeo, pillarMat);
    arm.position.set(-1, 3.5, 0);
    craneGroup.add(arm);
    portGroup.add(craneGroup);

    // Truck
    const truckGroup = new THREE.Group();

    // Truck Cabin
    const cabin3dGeo = new THREE.BoxGeometry(0.8, 0.9, 0.7);
    const cabin3dMat = new THREE.MeshStandardMaterial({ color: 0x3b82f6, metalness: 0.5 });
    const cabin3d = new THREE.Mesh(cabin3dGeo, cabin3dMat);
    cabin3d.position.set(1.2, 0.45, 0);
    cabin3d.castShadow = true;
    truckGroup.add(cabin3d);

    // Truck Trailer
    const trailerGeo = new THREE.BoxGeometry(2.2, 1, 0.8);
    const trailerMat = new THREE.MeshStandardMaterial({ color: 0xe2e8f0, metalness: 0.2 });
    const trailer = new THREE.Mesh(trailerGeo, trailerMat);
    trailer.position.set(-0.3, 0.6, 0);
    trailer.castShadow = true;
    truckGroup.add(trailer);

    // Wheels
    const wheelGeo = new THREE.CylinderGeometry(0.2, 0.2, 0.25, 12);
    const wheelMat = new THREE.MeshStandardMaterial({ color: 0x0f172a, roughness: 0.9 });

    const wheelPositions = [
      [1, 0.2, 0.4], [1, 0.2, -0.4],
      [-0.4, 0.2, 0.4], [-0.4, 0.2, -0.4],
      [-1, 0.2, 0.4], [-1, 0.2, -0.4]
    ];
    wheelPositions.forEach(([wx, wy, wz]) => {
      const wheel = new THREE.Mesh(wheelGeo, wheelMat);
      wheel.rotation.x = Math.PI / 2;
      wheel.position.set(wx, wy, wz);
      wheel.castShadow = true;
      truckGroup.add(wheel);
    });

    truckGroup.position.set(-1.5, 0.6, 0);
    portGroup.add(truckGroup);
    scene.add(portGroup);

    // --- Create 3D Cargo Plane ---
    const planeGroup = new THREE.Group();

    // Fuselage
    const fuseGeo = new THREE.CylinderGeometry(0.35, 0.2, 4, 16);
    const planeMat = new THREE.MeshStandardMaterial({ color: 0xffffff, metalness: 0.6, roughness: 0.2 });
    const fuselage = new THREE.Mesh(fuseGeo, planeMat);
    fuselage.rotation.x = Math.PI / 2;
    planeGroup.add(fuselage);

    // Wings
    const wingGeo = new THREE.BoxGeometry(4.5, 0.05, 0.8);
    const wing = new THREE.Mesh(wingGeo, planeMat);
    wing.position.set(0, 0, 0);
    planeGroup.add(wing);

    // Tail fin
    const tailGeo = new THREE.BoxGeometry(0.05, 0.8, 0.6);
    const tail = new THREE.Mesh(tailGeo, planeMat);
    tail.position.set(0, 0.4, -1.6);
    planeGroup.add(tail);

    scene.add(planeGroup);

    // --- Logistics Hubs & Connecting Arcs ---
    const hubs = [
      { pos: new THREE.Vector3(-6, 0.6, 4), color: 0x00bfa5 }, // Port
      { pos: new THREE.Vector3(4, 0, 2), color: 0x00bfa5 },    // Ship location
      { pos: new THREE.Vector3(6, 0, -6), color: 0x5a4fcf },   // Air hub
      { pos: new THREE.Vector3(-4, 0, -4), color: 0xf59e0b }   // Inland center
    ];

    hubs.forEach(hub => {
      const beaconGeo = new THREE.CylinderGeometry(0.01, 0.6, 0.8, 16, 1, true);
      const beaconMat = new THREE.MeshBasicMaterial({
        color: hub.color,
        transparent: true,
        opacity: 0.3,
        side: THREE.DoubleSide
      });
      const beacon = new THREE.Mesh(beaconGeo, beaconMat);
      beacon.position.copy(hub.pos);
      beacon.position.y += 0.4;
      scene.add(beacon);

      const coreGeo = new THREE.SphereGeometry(0.15, 16, 16);
      const coreMat = new THREE.MeshBasicMaterial({ color: hub.color });
      const core = new THREE.Mesh(coreGeo, coreMat);
      core.position.copy(hub.pos);
      core.position.y += 0.8;
      scene.add(core);
    });

    // Create curve lines (arcs)
    function createArc(p1, p2, height, color) {
      const points = [];
      const mid = new THREE.Vector3().addVectors(p1, p2).multiplyScalar(0.5);
      mid.y += height;

      const curve = new THREE.QuadraticBezierCurve3(p1, mid, p2);
      const curvePoints = curve.getPoints(30);

      const lineGeo = new THREE.BufferGeometry().setFromPoints(curvePoints);
      const lineMat = new THREE.LineBasicMaterial({ color: color, transparent: true, opacity: 0.4 });
      const arcLine = new THREE.Line(lineGeo, lineMat);
      scene.add(arcLine);

      // Return path and curve details for flow animation
      return { curve, length: curvePoints.length };
    }

    const arcs = [
      createArc(hubs[0].pos, hubs[2].pos, 6, 0x00bfa5), // Port → Air Hub
      createArc(hubs[3].pos, hubs[2].pos, 5, 0xf59e0b), // Inland → Air Hub
      createArc(hubs[3].pos, hubs[0].pos, 3, 0x3b82f6), // Inland → Port
    ];

    // Flowing particles along arcs
    const particlesGroup = new THREE.Group();
    const particleCount = 15;
    const particles = [];

    const particleGeo = new THREE.SphereGeometry(0.1, 8, 8);
    const particleMat = new THREE.MeshBasicMaterial({ color: 0x00bfa5 });

    for (let i = 0; i < particleCount; i++) {
      const p = new THREE.Mesh(particleGeo, particleMat);
      scene.add(p);
      particles.push({
        mesh: p,
        arc: arcs[i % arcs.length],
        progress: Math.random(),
        speed: 0.003 + Math.random() * 0.003
      });
    }

    // --- Interactive Mouse Tilt ---
    let mouseX = 0, mouseY = 0;
    const onMouseMove = (e) => {
      const rect = container.getBoundingClientRect();
      mouseX = ((e.clientX - rect.left) / rect.width) * 2 - 1;
      mouseY = -((e.clientY - rect.top) / rect.height) * 2 + 1;
    };
    container.addEventListener('mousemove', onMouseMove);

    // --- Resize handler ---
    const handleResize = () => {
      if (!container) return;
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    };
    window.addEventListener('resize', handleResize);

    // --- Animation loop ---
    let time = 0;
    let reqId;

    const animate = () => {
      reqId = requestAnimationFrame(animate);
      time += 0.01;

      // Animate cargo ship (gentle bobbing & sailing slightly)
      shipGroup.position.y = 0.4 + Math.sin(time * 2) * 0.05;
      shipGroup.rotation.z = Math.sin(time) * 0.02;
      shipGroup.rotation.y = Math.PI / 12 + Math.cos(time * 0.5) * 0.05;

      // Rotate crane cabin/arm
      craneGroup.rotation.y = Math.sin(time * 0.5) * 0.4;

      // Move truck back and forth
      truckGroup.position.x = -1.5 + Math.sin(time * 0.4) * 0.8;

      // Flight loop for plane (circular track in sky)
      const planeRadius = 11;
      const planeAngle = time * 0.2;
      planeGroup.position.set(
        Math.cos(planeAngle) * planeRadius,
        7 + Math.sin(time * 0.5) * 0.5,
        Math.sin(planeAngle) * planeRadius
      );
      planeGroup.rotation.y = -planeAngle + Math.PI / 2;
      planeGroup.rotation.z = 0.15; // bank angle

      // Animate flowing cargo particles along arcs
      particles.forEach(p => {
        p.progress += p.speed;
        if (p.progress > 1) p.progress = 0;
        const pos = p.arc.curve.getPointAt(p.progress);
        p.mesh.position.copy(pos);
      });

      // Mouse interactive tilt camera influence
      const targetCamX = 15 + mouseX * 3;
      const targetCamY = 12 + mouseY * 2;
      camera.position.x += (targetCamX - camera.position.x) * 0.05;
      camera.position.y += (targetCamY - camera.position.y) * 0.05;
      camera.lookAt(0, 1, 0);

      renderer.render(scene, camera);
    };

    animate();

    // --- Cleanup ---
    return () => {
      cancelAnimationFrame(reqId);
      window.removeEventListener('resize', handleResize);
      if (container) {
        container.removeEventListener('mousemove', onMouseMove);
        if (container.contains(renderer.domElement)) {
          container.removeChild(renderer.domElement);
        }
      }
      scene.clear();
      renderer.dispose();
    };
  }, []);

  return (
    <div
      ref={containerRef}
      style={{
        width: '100%',
        height: '100%',
        minHeight: '380px',
        position: 'relative',
        overflow: 'hidden',
        borderRadius: '20px',
      }}
    />
  );
}
