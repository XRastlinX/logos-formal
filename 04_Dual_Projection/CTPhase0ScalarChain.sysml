package CTPhase0ScalarChain {
    
    // Morphism f
    calc def AltitudeToPressure {
        in attribute altitude :> ISQ::length;
        in attribute referencePressure :> ISQ::pressure;
        return :> ISQ::pressure =
            (1 - altitude / 10[km]) * referencePressure;
    }

    // Morphism g
    calc def PressureToDrag {
        in attribute pressure :> ISQ::pressure;
        in attribute referencePressure :> ISQ::pressure;
        in attribute referenceDrag :> ISQ::force;
        return :> ISQ::force =
            (referenceDrag / referencePressure) * pressure;
    }
    
    // Composition g ∘ f
    calc def AltitudeToDrag {
        in attribute altitude :> ISQ::length;
        in attribute referencePressure :> ISQ::pressure;
        in attribute referenceDrag :> ISQ::force;
        return :> ISQ::force = 
            PressureToDrag(AltitudeToPressure(altitude, referencePressure), referencePressure, referenceDrag);
    }

    // Carrier Set Constraints (Predicates in Rel)
    constraint def PilotAltitudeDomain {
        in attribute altitude :> ISQ::length;
        altitude >= 0[km] and altitude <= 5[km]
    }
    
    constraint def PilotPressureDomain {
        in attribute pressure :> ISQ::pressure;
        in attribute referencePressure :> ISQ::pressure;
        pressure >= 0.5 * referencePressure and pressure <= referencePressure
    }

    constraint def PilotDragDomain {
        in attribute drag :> ISQ::force;
        in attribute referenceDrag :> ISQ::force;
        drag >= 0.5 * referenceDrag and drag <= referenceDrag
    }
}
