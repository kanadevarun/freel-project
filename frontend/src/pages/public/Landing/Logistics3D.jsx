import React, { useEffect, useRef } from 'react';
import * as THREE from 'three';

export function Logistics3D() {
  const containerRef = useRef(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    // --- 1. Scene Setup ---
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(0xf4f7f6);
    scene.fog = new THREE.FogExp2(0xf4f7f6, 0.012);

    // --- 2. Camera Rig Setup (Crucial for combining Animation + Mouse Parallax) ---
    const cameraGroup = new THREE.Group();
    scene.add(cameraGroup);

    const camera = new THREE.PerspectiveCamera(45, container.clientWidth / container.clientHeight, 0.1, 1000);
    cameraGroup.add(camera);

    // Using a Vector3 to keep track of the look-at target
    const camTarget = new THREE.Vector3(0, 3, 0);
    cameraGroup.position.set(0, 25, 35); // Initial Wide position

    // --- 3. Renderer ---
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    container.appendChild(renderer.domElement);

    // --- 4. Lighting (Clean, Explainer-Video Style) ---
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.7);
    scene.add(ambientLight);

    const dirLight = new THREE.DirectionalLight(0xffffff, 0.9);
    dirLight.position.set(20, 40, 20);
    dirLight.castShadow = true;
    dirLight.shadow.mapSize.width = 2048;
    dirLight.shadow.mapSize.height = 2048;
    dirLight.shadow.camera.near = 0.5;
    dirLight.shadow.camera.far = 100;
    dirLight.shadow.camera.left = -25;
    dirLight.shadow.camera.right = 25;
    dirLight.shadow.camera.top = 25;
    dirLight.shadow.camera.bottom = -25;
    dirLight.shadow.bias = -0.0005;
    scene.add(dirLight);

    const fillLight = new THREE.PointLight(0x0ea5e9, 0.4, 100);
    fillLight.position.set(-20, 10, -20);
    scene.add(fillLight);

    // --- 5. Materials ---
    const matFloor = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: 0.1 });
    const matRack = new THREE.MeshStandardMaterial({ color: 0x334155, roughness: 0.6 });
    const matBoxTeal = new THREE.MeshStandardMaterial({ color: 0x0ea5e9, roughness: 0.3 });
    const matBoxOrange = new THREE.MeshStandardMaterial({ color: 0xf97316, roughness: 0.3 });
    const matBoxWhite = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: 0.2 });
    const matBelt = new THREE.MeshStandardMaterial({ color: 0x1e293b, roughness: 0.8 });
    const matTruckCab = new THREE.MeshStandardMaterial({ color: 0x0ea5e9, roughness: 0.2 });
    const matTire = new THREE.MeshStandardMaterial({ color: 0x0f172a, roughness: 0.9 });
    const boxMaterials = [matBoxTeal, matBoxOrange, matBoxWhite];

    // --- 6. Environment Construction (Left-to-Right Workflow) ---
    const logisticsWorld = new THREE.Group();
    scene.add(logisticsWorld);

    // Floor
    const floor = new THREE.Mesh(new THREE.PlaneGeometry(200, 200), matFloor);
    floor.rotation.x = -Math.PI / 2;
    floor.receiveShadow = true;
    logisticsWorld.add(floor);

    // A. STORAGE (Left Side: x = -12 to -4)
    const createRack = (x, z) => {
      const rack = new THREE.Group();
      const beamGeo = new THREE.BoxGeometry(0.2, 8, 0.2);

      [1.1, -1.1].forEach(zOffset => {
        [-3, 0, 3].forEach(xOffset => {
          const beam = new THREE.Mesh(beamGeo, matRack);
          beam.position.set(xOffset, 4, zOffset);
          beam.castShadow = true;
          rack.add(beam);
        });
      });

      for (let i = 1; i <= 3; i++) {
        const shelf = new THREE.Mesh(new THREE.BoxGeometry(6.4, 0.1, 2.4), matRack);
        shelf.position.set(0, i * 2, 0);
        shelf.castShadow = true; shelf.receiveShadow = true;
        rack.add(shelf);

        // Populate Shelf with Boxes
        for (let b = -2; b <= 2; b++) {
          if (Math.random() > 0.4) {
            const size = 0.8 + Math.random() * 0.3;
            const box = new THREE.Mesh(new THREE.BoxGeometry(size, size, size), boxMaterials[Math.floor(Math.random() * 3)]);
            box.position.set(b * 1.2 + (Math.random() * 0.2), (i * 2) + (size / 2) + 0.05, (Math.random() * 0.4 - 0.2));
            box.castShadow = true;
            rack.add(box);
          }
        }
      }
      rack.position.set(x, 0, z);
      return rack;
    };

    logisticsWorld.add(createRack(-12, -6));
    logisticsWorld.add(createRack(-12, 2));

    // B. CONVEYOR BELT (Middle: x = -2 to 6)
    const beltGroup = new THREE.Group();
    const belt = new THREE.Mesh(new THREE.BoxGeometry(16, 0.2, 1.5), matBelt);
    belt.position.set(2, 1.5, 0);
    belt.receiveShadow = true; belt.castShadow = true;
    beltGroup.add(belt);

    const beltLegGeo = new THREE.CylinderGeometry(0.1, 0.1, 1.5);
    for (let i = -5; i <= 9; i += 3) {
      const leg1 = new THREE.Mesh(beltLegGeo, matRack); leg1.position.set(i, 0.75, 0.6);
      const leg2 = new THREE.Mesh(beltLegGeo, matRack); leg2.position.set(i, 0.75, -0.6);
      beltGroup.add(leg1, leg2);
    }
    logisticsWorld.add(beltGroup);

    // Moving Boxes on Conveyor
    const movingBoxes = Array.from({ length: 4 }).map((_, i) => {
      const box = new THREE.Mesh(new THREE.BoxGeometry(1, 1, 1), matBoxOrange);
      box.castShadow = true;
      box.position.set(-6 + (i * 4), 2.1, 0); // Start along belt
      logisticsWorld.add(box);
      return box;
    });

    // C. TRUCK LOADING (Right Side: x = 12)
    const truck = new THREE.Group();
    const trailer = new THREE.Mesh(new THREE.BoxGeometry(7, 4, 3), matBoxWhite);
    trailer.position.set(0, 3, 0);
    trailer.castShadow = true;
    const cab = new THREE.Mesh(new THREE.BoxGeometry(2.5, 3, 3), matTruckCab);
    cab.position.set(4.75, 2.5, 0);
    cab.castShadow = true;
    truck.add(trailer, cab);

    const wheelGeo = new THREE.CylinderGeometry(0.7, 0.7, 0.5, 16);
    wheelGeo.rotateX(Math.PI / 2);
    [[-2.5, 0.7, 1.5], [-2.5, 0.7, -1.5], [2.5, 0.7, 1.5], [2.5, 0.7, -1.5], [4.75, 0.7, 1.5], [4.75, 0.7, -1.5]].forEach(pos => {
      const wheel = new THREE.Mesh(wheelGeo, matTire);
      wheel.position.set(...pos);
      wheel.castShadow = true;
      truck.add(wheel);
    });
    truck.position.set(13, 0, 0);
    logisticsWorld.add(truck);

    // D. DRONE (Hovering above Truck)
    const drone = new THREE.Group();
    const droneBody = new THREE.Mesh(new THREE.BoxGeometry(1.2, 0.4, 1.2), matBoxWhite);
    droneBody.castShadow = true;
    drone.add(droneBody);

    const props = [];
    [[0.7, 0.7], [-0.7, 0.7], [0.7, -0.7], [-0.7, -0.7]].forEach(pos => {
      const arm = new THREE.Mesh(new THREE.CylinderGeometry(0.05, 0.05, 0.8), matRack);
      arm.rotation.x = Math.PI / 2;
      arm.position.set(pos / 2, 0, pos[1] / 2);
      const prop = new THREE.Mesh(new THREE.BoxGeometry(1.4, 0.05, 0.1), matBoxTeal);
      prop.position.set(pos, 0.2, pos[1]);
      prop.castShadow = true;
      drone.add(arm, prop);
      props.push(prop);
    });
    const dronePack = new THREE.Mesh(new THREE.BoxGeometry(0.8, 0.8, 0.8), matBoxOrange);
    dronePack.position.set(0, -0.6, 0);
    dronePack.castShadow = true;
    drone.add(dronePack);
    drone.position.set(13, 9, 3);
    logisticsWorld.add(drone);

    // --- 7. Native Cinematic Camera Animation ---
    // Define the keyframes for the camera path and look target
    const keyframes = [
      { pos: new THREE.Vector3(0, 25, 35), target: new THREE.Vector3(0, 3, 0) }, // Wide
      { pos: new THREE.Vector3(-12, 10, 15), target: new THREE.Vector3(-12, 3, 0) }, // Storage
      { pos: new THREE.Vector3(2, 8, 12), target: new THREE.Vector3(2, 2, 0) }, // Belt
      { pos: new THREE.Vector3(13, 10, 15), target: new THREE.Vector3(13, 4, 0) }, // Truck
      { pos: new THREE.Vector3(0, 25, 35), target: new THREE.Vector3(0, 3, 0) } // Wide
    ];

    let currentPhase = 0;
    let transitionProgress = 0;
    const phaseDuration = 7; // seconds per move
    const stayDuration = 1.5; // seconds to stay at location
    let isStaying = false;
    let stayTimer = 0;
    let lastTime = 0;

    // Helper for smooth easing (easeInOutCubic)
    const easeInOutCubic = (t) => t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;

    // Set initial positions
    cameraGroup.position.copy(keyframes[0].pos);
    camTarget.copy(keyframes[0].target);

    // --- 8. Mouse Parallax Interaction ---
    let mouseX = 0, mouseY = 0;
    const targetMouse = { x: 0, y: 0 };

    const onMouseMove = (e) => {
      // Normalize mouse to -1 to 1
      targetMouse.x = (e.clientX / window.innerWidth) * 2 - 1;
      targetMouse.y = -(e.clientY / window.innerHeight) * 2 + 1;
    };
    window.addEventListener('mousemove', onMouseMove);

    // --- 9. Render Loop ---
    let reqId;
    const clock = new THREE.Clock();

    const animate = () => {
      reqId = requestAnimationFrame(animate);

      const time = clock.getElapsedTime();
      const delta = time - lastTime;
      lastTime = time;

      // Animate Conveyor Boxes
      movingBoxes.forEach(box => {
        box.position.x += 0.05; // Move right towards truck
        if (box.position.x > 9) box.position.x = -6; // Reset to start of belt
      });

      // Animate Drone Hover & Propellers
      drone.position.y = 9 + Math.sin(time * 2) * 0.4;
      drone.rotation.y = Math.sin(time * 0.5) * 0.15;
      props.forEach(prop => { prop.rotation.y -= 0.5; });

      // --- Execute Cinematic Camera Logic ---
      if (isStaying) {
        stayTimer += delta;
        if (stayTimer >= stayDuration) {
          isStaying = false;
          stayTimer = 0;
          currentPhase = (currentPhase + 1) % (keyframes.length - 1);
        }
      } else {
        transitionProgress += (delta / phaseDuration);

        if (transitionProgress >= 1) {
          transitionProgress = 1;
          isStaying = true;
        }

        const startFrame = keyframes[currentPhase];
        const endFrame = keyframes[currentPhase + 1];

        const easedT = easeInOutCubic(transitionProgress);

        cameraGroup.position.lerpVectors(startFrame.pos, endFrame.pos, easedT);
        camTarget.lerpVectors(startFrame.target, endFrame.target, easedT);

        if (transitionProgress === 1) {
          transitionProgress = 0;
        }
      }

      // Smooth Mouse Parallax (Damped)
      mouseX += (targetMouse.x - mouseX) * 0.05;
      mouseY += (targetMouse.y - mouseY) * 0.05;

      // Apply parallax to internal camera relative to animated cameraGroup
      camera.position.x = mouseX * 2;
      camera.position.y = mouseY * 2;

      // Ensure camera always looks at the animated target, adjusted by mouse
      camera.lookAt(
        camTarget.x - mouseX * 2,
        camTarget.y - mouseY * 2,
        camTarget.z
      );

      renderer.render(scene, camera);
    };

    // Start animation loop
    animate();

    // --- 10. Responsive & Cleanup ---
    const handleResize = () => {
      if (!container) return;
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    };
    window.addEventListener('resize', handleResize);

    return () => {
      cancelAnimationFrame(reqId);
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('mousemove', onMouseMove);
      if (container && container.contains(renderer.domElement)) {
        container.removeChild(renderer.domElement);
      }
      scene.clear();
      renderer.dispose();
    };
  }, []);

  // Important: pointerEvents: 'none' ensures the 3D canvas doesn't block UI clicks!
  return <div ref={containerRef} style={{ width: '100%', height: '100%', pointerEvents: 'none' }} />;
}

export default Logistics3D;